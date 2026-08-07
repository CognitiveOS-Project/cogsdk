package daemon

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"
)

// Envelope is the wire format exchanged with the cognitiveosd Unix socket.
//
// It matches the authoritative definition in cognitiveosd/internal/daemon.
// Payload holds the message-specific JSON payload as raw bytes; decode it
// with json.Unmarshal into the matching payload type.
type Envelope struct {
	Type      string          `json:"type"`
	ID        string          `json:"id,omitempty"`
	Timestamp string          `json:"timestamp,omitempty"`
	From      string          `json:"from,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

// NewEnvelope builds an Envelope with a fresh ID and UTC timestamp.
func NewEnvelope(msgType, from string, payload interface{}) Envelope {
	b, _ := json.Marshal(payload)
	return Envelope{
		Type:      msgType,
		ID:        newID(),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		From:      from,
		Payload:   b,
	}
}

// PayloadStatus is the standard response envelope payload: status is "ok" or
// "error"; on error, Error carries the code and message.
type PayloadStatus struct {
	Status string     `json:"status"`
	Error  *ErrorInfo `json:"error,omitempty"`
	Data   any        `json:"data,omitempty"`
}

// ErrorInfo carries an error code and human-readable message.
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorPayload builds a status payload with Status "error".
func ErrorPayload(code, message string) PayloadStatus {
	return PayloadStatus{Status: "error", Error: &ErrorInfo{Code: code, Message: message}}
}

// OKPayload builds a status payload with Status "ok" and optional data.
func OKPayload(data any) PayloadStatus {
	return PayloadStatus{Status: "ok", Data: data}
}

// newID returns a random UUID v4 string, mirroring cognitiveosd's uuidV4.
func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
