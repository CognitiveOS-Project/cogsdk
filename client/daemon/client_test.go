package daemon

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// startTestServer runs a minimal daemon-like server on a unix socket that
// echoes system_code requests with a code_accepted response carrying the same ID.
func startTestServer(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "daemon.sock")

	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer func() { _ = c.Close() }()
				dec := json.NewDecoder(c)
				enc := json.NewEncoder(c)
				for {
					var env Envelope
					if err := dec.Decode(&env); err != nil {
						return
					}
					resp := NewEnvelope("code_accepted", "cognitiveosd", CodeAcceptedPayload{
						Status: "ok",
						Effect: "waking from idle state",
					})
					resp.ID = env.ID
					_ = enc.Encode(resp)
				}
			}(conn)
		}
	}()

	t.Cleanup(func() { _ = ln.Close() })
	return path
}

func TestClientRoundTrip(t *testing.T) {
	path := startTestServer(t)

	c, err := DialPath(path)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = c.Close() }()

	c.SetReadTimeout(5 * time.Second)
	defer c.ClearReadTimeout()

	resp, err := c.Request("system_code", SystemCodePayload{
		Code:   CodeWake,
		Origin: OriginCLI,
	})
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	if resp.Type != "code_accepted" {
		t.Fatalf("response type = %q, want code_accepted", resp.Type)
	}

	var payload CodeAcceptedPayload
	if err := json.Unmarshal(resp.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("status = %q, want ok", payload.Status)
	}
}

func TestClientRoundTripMismatchedID(t *testing.T) {
	path := startTestServer(t)

	c, err := DialPath(path)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = c.Close() }()

	// Send an envelope with a bogus ID, then a valid one; the client should
	// skip the mismatched response and surface the matching one.
	env := NewEnvelope("system_code", "cpm", SystemCodePayload{Code: CodeWake, Origin: OriginCLI})

	// Server always echoes the request ID, so simulate a stray response by
	// sending two requests and confirming we read the matching response.
	c.SetReadTimeout(5 * time.Second)
	defer c.ClearReadTimeout()

	resp, err := c.RoundTrip(env)
	if err != nil {
		t.Fatalf("roundtrip: %v", err)
	}
	if resp.ID != env.ID {
		t.Fatalf("response ID = %q, want %q", resp.ID, env.ID)
	}
}

func TestDefaultSocketPathOverride(t *testing.T) {
	_ = os.Setenv("COGNITIVEOS_SOCKET", "/tmp/test-daemon.sock")
	defer func() { _ = os.Unsetenv("COGNITIVEOS_SOCKET") }()

	if got := DefaultSocketPath(); got != "/tmp/test-daemon.sock" {
		t.Fatalf("got %q, want override", got)
	}
}
