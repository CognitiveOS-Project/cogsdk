package daemon

import "encoding/json"

// Envelope is the wire format exchanged with the cognitiveosd Unix socket.
type Envelope struct {
	Type    string          `json:"type"`
	From    string          `json:"from"`
	Payload json.RawMessage `json:"payload"`
}
