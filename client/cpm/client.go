/*
Package cpm provides the typed CPM RPC client.

Per ADR-011, it is the only path to nested .cgp tools: tool invocation is
brokered by the cpm daemon (ADR-004 validated tool domains), and cogsdk ships
typed clients only — it never executes tools directly.

The client speaks the cognitiveosd socket protocol (client/daemon) for the
package-manager message surface (cpm_tune and friends).
*/
package cpm

import (
	"fmt"

	"github.com/CognitiveOS-Project/cogsdk/client/daemon"
)

// Client is a CPM RPC client. It embeds the underlying daemon socket client.
type Client struct {
	dc *daemon.Client
}

// Dial connects to the cognitiveosd socket as the CPM component.
func Dial() (*Client, error) {
	dc, err := daemon.Dial()
	if err != nil {
		return nil, err
	}
	dc.SetFrom("cpm")
	return &Client{dc: dc}, nil
}

// New wraps an existing daemon client as a CPM client.
func New(dc *daemon.Client) *Client {
	return &Client{dc: dc}
}

// Close closes the underlying daemon socket connection.
func (c *Client) Close() error {
	return c.dc.Close()
}

// TunePayload is the payload of the cpm_tune message.
type TunePayload struct {
	Package    string `json:"package"`
	Background bool   `json:"background"`
	Epochs     int    `json:"epochs"`
	Finalize   bool   `json:"finalize"`
	Quantize   string `json:"quantize"`
}

// TuneOptions configures a tuning request.
type TuneOptions struct {
	Background bool
	Epochs     int
	Finalize   bool
	Quantize   string
}

// Tune triggers background tuning of an installed package via the daemon.
func (c *Client) Tune(pkg string, opts TuneOptions) error {
	if pkg == "" {
		return fmt.Errorf("package name is required")
	}
	_, err := c.dc.Request("cpm_tune", TunePayload{
		Package:    pkg,
		Background: opts.Background,
		Epochs:     opts.Epochs,
		Finalize:   opts.Finalize,
		Quantize:   opts.Quantize,
	})
	return err
}
