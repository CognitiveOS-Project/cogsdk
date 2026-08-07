package daemon

import (
	"encoding/json"
	"testing"
)

func TestEnvelopeRoundTrip(t *testing.T) {
	in := Envelope{
		Type:    "test",
		From:    "cogsdk",
		Payload: json.RawMessage(`{"n":1}`),
	}

	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out Envelope
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if out.Type != in.Type || out.From != in.From {
		t.Fatalf("got %+v, want %+v", out, in)
	}
}
