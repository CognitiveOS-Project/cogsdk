package daemon

import (
	"encoding/json"
	"testing"
)

func TestEnvelopeRoundTrip(t *testing.T) {
	env := NewEnvelope("test", "cogsdk", map[string]string{"n": "1"})
	if env.Type != "test" {
		t.Fatalf("Type = %q, want %q", env.Type, "test")
	}
	if env.From != "cogsdk" {
		t.Fatalf("From = %q, want %q", env.From, "cogsdk")
	}
	if env.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out Envelope
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Type != env.Type || out.ID != env.ID {
		t.Fatalf("got %+v, want %+v", out, env)
	}

	var payload map[string]string
	if err := json.Unmarshal(out.Payload, &payload); err != nil {
		t.Fatalf("payload: %v", err)
	}
	if payload["n"] != "1" {
		t.Fatalf("payload = %v", payload)
	}
}

func TestPayloadStatusError(t *testing.T) {
	p := ErrorPayload("E_TEST", "boom")
	if p.Status != "error" || p.Error == nil || p.Error.Code != "E_TEST" || p.Error.Message != "boom" {
		t.Fatalf("unexpected error payload: %+v", p)
	}
}

func TestPayloadStatusOK(t *testing.T) {
	p := OKPayload(map[string]int{"a": 1})
	if p.Status != "ok" || p.Error != nil {
		t.Fatalf("unexpected ok payload: %+v", p)
	}
}

func TestValidCode(t *testing.T) {
	for _, c := range []string{CodeWake, CodeIdle, CodeSecurity, CodeReset, CodeUnlock} {
		if !ValidCode(c) {
			t.Errorf("expected %q to be valid", c)
		}
	}
	if ValidCode("bogus") {
		t.Error("expected bogus to be invalid")
	}
}
