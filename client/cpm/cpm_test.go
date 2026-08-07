package cpm

import (
	"encoding/json"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/CognitiveOS-Project/cogsdk/client/daemon"
)

// startDaemonServer mimics cognitiveosd's handling of a cpm_tune envelope.
func startDaemonServer(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "daemon.sock")

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
				defer c.Close()
				dec := json.NewDecoder(c)
				enc := json.NewEncoder(c)
				for {
					var env daemon.Envelope
					if err := dec.Decode(&env); err != nil {
						return
					}
					resp := daemon.NewEnvelope("tune_accepted", "cognitiveosd", daemon.OKPayload(nil))
					resp.ID = env.ID
					enc.Encode(resp)
				}
			}(conn)
		}
	}()

	t.Cleanup(func() { ln.Close() })
	return path
}

func TestTune(t *testing.T) {
	path := startDaemonServer(t)

	dc, err := daemon.DialPath(path)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer dc.Close()

	dc.SetReadTimeout(5 * time.Second)
	defer dc.ClearReadTimeout()

	c := New(dc)
	err = c.Tune("org/my-model", TuneOptions{
		Background: true,
		Epochs:     3,
		Finalize:   true,
		Quantize:   "q4_0",
	})
	if err != nil {
		t.Fatalf("tune: %v", err)
	}
}

func TestNewDoesNotMutateFrom(t *testing.T) {
	dc := &daemon.Client{}
	dc.SetFrom("anything")
	c := New(dc)
	if c.dc.From() != "anything" {
		t.Fatalf("New must not mutate From, got %q", c.dc.From())
	}
}
