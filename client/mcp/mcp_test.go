package mcp

import (
	"encoding/json"
	"errors"
	"net"
	"testing"
)

// TestServerClientRoundTrip wires a Server and a Client over an in-memory pipe.
func TestServerClientRoundTrip(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer func() { _ = serverConn.Close() }()
	defer func() { _ = clientConn.Close() }()

	srv := NewWithVersion("test-bridge", "1.0.0")
	srv.Tools = []Tool{{
		Name:        "cognitiveos.test.echo",
		Description: "echo text",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"text": map[string]interface{}{"type": "string"},
			},
			"required": []interface{}{"text"},
		},
	}}
	srv.Handle("cognitiveos.test.echo", func(args map[string]interface{}) (interface{}, error) {
		return args["text"], nil
	})

	go func() { _ = srv.Run(serverConn, serverConn) }()

	client := NewClient(clientConn, clientConn)

	tools, err := client.ListTools()
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(tools) != 1 || tools[0].Name != "cognitiveos.test.echo" {
		t.Fatalf("tools = %+v", tools)
	}

	res, err := client.Call("cognitiveos.test.echo", map[string]interface{}{"text": "hello"})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error result: %+v", res)
	}
	if res.Text() != "hello" {
		t.Fatalf("text = %q, want hello", res.Text())
	}
}

func TestServerToolNotFound(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer func() { _ = serverConn.Close() }()
	defer func() { _ = clientConn.Close() }()

	srv := New("empty-bridge")
	go func() { _ = srv.Run(serverConn, serverConn) }()

	client := NewClient(clientConn, clientConn)
	res, err := client.Call("cognitiveos.nope.nope", nil)
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected error result, got %+v", res)
	}
	code, _ := ParseErrorText(res.Text())
	if code != "E_NOT_FOUND" {
		t.Fatalf("code = %q, want E_NOT_FOUND", code)
	}
}

func TestServerHandlerError(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer func() { _ = serverConn.Close() }()
	defer func() { _ = clientConn.Close() }()

	srv := New("error-bridge")
	srv.Handle("cognitiveos.test.fail", func(args map[string]interface{}) (interface{}, error) {
		return nil, errors.New("boom")
	})

	go func() { _ = srv.Run(serverConn, serverConn) }()

	client := NewClient(clientConn, clientConn)
	res, err := client.Call("cognitiveos.test.fail", nil)
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected error result")
	}
	code, _ := ParseErrorText(res.Text())
	if code != "E_INTERNAL" {
		t.Fatalf("code = %q, want E_INTERNAL", code)
	}
}

func TestParseErrorText(t *testing.T) {
	code, msg := ParseErrorText("ERROR:E_TIMEOUT: operation timed out")
	if code != "E_TIMEOUT" || msg != "operation timed out" {
		t.Fatalf("got %q %q", code, msg)
	}
}

func TestServerHealthcheck(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer func() { _ = serverConn.Close() }()
	defer func() { _ = clientConn.Close() }()

	srv := New("health-bridge")
	go func() { _ = srv.Run(serverConn, serverConn) }()

	// Send the healthcheck notification; the server replies with
	// {"type":"healthcheck_ok",...} on the same writer.
	req := rpcRequest{JSONRPC: JSONRPCVersion, Method: "healthcheck"}
	if err := json.NewEncoder(clientConn).Encode(req); err != nil {
		t.Fatalf("encode: %v", err)
	}

	dec := json.NewDecoder(clientConn)
	var resp map[string]interface{}
	if err := dec.Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["type"] != "healthcheck_ok" {
		t.Fatalf("type = %v", resp["type"])
	}
}
