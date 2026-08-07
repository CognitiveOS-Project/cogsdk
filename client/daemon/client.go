package daemon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

// DefaultSocketPath returns the cognitiveosd socket path, honoring the
// COGNITIVEOS_SOCKET environment override.
func DefaultSocketPath() string {
	if p := os.Getenv("COGNITIVEOS_SOCKET"); p != "" {
		return p
	}
	return "/cognitiveos/run/daemon.sock"
}

// Client is a JSON-over-Unix-socket client for cognitiveosd.
//
// Messages are newline-delimited JSON Envelopes (see socket.go in
// cognitiveosd/internal/daemon). SendMessage writes a single envelope;
// RoundTrip writes one and reads the daemon's matching response.
type Client struct {
	conn    net.Conn
	scanner *bufio.Scanner
	from    string
}

// Dial connects to the cognitiveosd socket (see DefaultSocketPath).
func Dial() (*Client, error) {
	return DialPath(DefaultSocketPath())
}

// DialPath connects to an explicit socket path.
func DialPath(path string) (*Client, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return nil, fmt.Errorf("daemon connection failed: %w", err)
	}
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 3*1024*1024), 3*1024*1024)
	return &Client{conn: conn, scanner: scanner}, nil
}

// SetFrom sets the From field used by outgoing envelopes.
func (c *Client) SetFrom(from string) {
	c.from = from
}

// From returns the From field used by outgoing envelopes.
func (c *Client) From() string {
	return c.from
}

// Close closes the underlying connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// SendMessage marshals payload into an Envelope and writes it, fire-and-forget.
func (c *Client) SendMessage(msgType string, payload interface{}) error {
	env := NewEnvelope(msgType, c.from, payload)
	return c.SendEnvelope(env)
}

// SendEnvelope writes a single envelope as newline-delimited JSON.
func (c *Client) SendEnvelope(env Envelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal envelope failed: %w", err)
	}
	if _, err := c.conn.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write envelope failed: %w", err)
	}
	return nil
}

// RoundTrip sends an envelope and reads envelopes until one arrives whose ID
// matches the sent envelope's ID (the daemon echoes the request ID). Returns
// the matching response envelope.
func (c *Client) RoundTrip(env Envelope) (Envelope, error) {
	if err := c.SendEnvelope(env); err != nil {
		return Envelope{}, err
	}
	return c.ReadResponse(env.ID)
}

// ReadResponse reads envelopes until one arrives whose ID matches wantID.
func (c *Client) ReadResponse(wantID string) (Envelope, error) {
	for c.scanner.Scan() {
		var resp Envelope
		if err := json.Unmarshal(c.scanner.Bytes(), &resp); err != nil {
			return Envelope{}, fmt.Errorf("parse envelope: %w", err)
		}
		if wantID != "" && resp.ID != wantID {
			continue
		}
		return resp, nil
	}
	if err := c.scanner.Err(); err != nil {
		return Envelope{}, fmt.Errorf("read envelope: %w", err)
	}
	return Envelope{}, fmt.Errorf("connection closed before response")
}

// Request sends a typed request and returns the daemon's response envelope.
// It returns the response even when the daemon reports an error status; callers
// should check PayloadStatus.Status.
func (c *Client) Request(msgType string, payload interface{}) (Envelope, error) {
	return c.RoundTrip(NewEnvelope(msgType, c.from, payload))
}

// SetReadTimeout sets a read deadline for RoundTrip/ReadResponse calls.
func (c *Client) SetReadTimeout(d time.Duration) {
	c.conn.SetReadDeadline(time.Now().Add(d))
}

// ClearReadTimeout removes any read deadline.
func (c *Client) ClearReadTimeout() {
	c.conn.SetReadDeadline(time.Time{})
}
