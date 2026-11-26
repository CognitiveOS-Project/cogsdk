// Package cograw provides a client for the cograw Raw Model RPC server.
//
// cograw is a root-level RPC server that communicates over a Unix socket using
// JSON-RPC 2.0. The CograwClient provides methods to validate system codes,
// unlock codes, audit resources, validate prompts, and validate package requests.
//
// All communication is over a Unix socket at the default path
// /cognitiveos/run/raw.sock, or via COGNITIVEOS_RAW_SOCKET environment variable.
package cograw

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
)

// DefaultSocketPath returns the cograw socket path, honoring the
// COGNITIVEOS_RAW_SOCKET environment override.
func DefaultSocketPath() string {
	if p := os.Getenv("COGNITIVEOS_RAW_SOCKET"); p != "" {
		return p
	}
	return "/cognitiveos/run/raw.sock"
}

// RPCRequest is a JSON-RPC 2.0 request to cograw.
type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// RPCResponse is a JSON-RPC 2.0 response from cograw.
type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError is a JSON-RPC 2.0 error object.
type RPCError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ValidationParams are parameters for the validate_package_request RPC method.
type ValidationParams struct {
	Operation        string                   `json:"operation"`
	PackageName      string                   `json:"package_name"`
	Version          string                   `json:"version,omitempty"`
	ManifestMetadata *PackageManifestMetadata `json:"manifest_metadata,omitempty"`
}

// PackageManifestMetadata is metadata about a package manifest.
type PackageManifestMetadata struct {
	HasRawModel bool   `json:"has_raw_model,omitempty"`
	DiskSpaceMB int64  `json:"disk_space_mb,omitempty"`
	Registry    string `json:"registry,omitempty"`
	IsCritical  bool   `json:"is_critical,omitempty"`
}

// ValidationResult is the result of a package validation request.
type ValidationResult struct {
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Command string `json:"command"`
}

// Client is a JSON-RPC 2.0 client for the cograw Raw Model server.
//
// CograwClient communicates over a Unix socket using newline-delimited JSON-RPC
// 2.0 envelopes. All methods are safe for concurrent use.
type Client struct {
	conn    net.Conn
	scanner *bufio.Scanner
	mu      sync.Mutex
	id      int
	ready   bool
}

// NewClient creates a new CograwClient connected to the given socket path.
// If path is empty, DefaultSocketPath() is used.
func NewClient(path string) (*Client, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return nil, fmt.Errorf("cograw connection failed: %w", err)
	}
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 3*1024*1024), 3*1024*1024)
	return &Client{
		conn:    conn,
		scanner: scanner,
	}, nil
}

// Dial connects to the default cograw socket path.
func Dial() (*Client, error) {
	return NewClient(DefaultSocketPath())
}

// DialPath connects to an explicit socket path.
func DialPath(path string) (*Client, error) {
	return NewClient(path)
}

// Connect establishes a connection and verifies cograw is ready via healthcheck.
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("cograw not connected")
	}

	health, err := c.call("healthcheck", nil)
	if err != nil {
		return fmt.Errorf("raw healthcheck: %w", err)
	}

	var healthResult struct {
		Status      string `json:"status"`
		ModelLoaded bool   `json:"model_loaded"`
	}
	if err := json.Unmarshal(health, &healthResult); err != nil {
		return fmt.Errorf("parse health result: %w", err)
	}

	if healthResult.Status != "ready" {
		return fmt.Errorf("raw model not ready: %s", healthResult.Status)
	}

	c.ready = true
	return nil
}

// ValidateSystemCode validates a system code via cograw.
func (c *Client) ValidateSystemCode(code, origin string) (string, string, error) {
	params := map[string]string{
		"code":   code,
		"origin": origin,
	}

	result, err := c.call("validate_system_code", params)
	if err != nil {
		return "", "", err
	}

	var resp struct {
		Status string `json:"status"`
		Action string `json:"action"`
	}
	if err := json.Unmarshal(result, &resp); err != nil {
		return "", "", fmt.Errorf("parse validate response: %w", err)
	}

	return resp.Status, resp.Action, nil
}

// CheckUnlockCode checks an unlock code via cograw.
func (c *Client) CheckUnlockCode(code, patchName string) (string, string, error) {
	params := map[string]string{
		"code":       code,
		"patch_name": patchName,
	}

	result, err := c.call("check_unlock_code", params)
	if err != nil {
		return "", "", err
	}

	var resp struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(result, &resp); err != nil {
		return "", "", fmt.Errorf("parse unlock response: %w", err)
	}

	return resp.Status, resp.Message, nil
}

// AuditResources audits hardware resources via cograw.
func (c *Client) AuditResources(requestedMB int64) (bool, int64, int64, bool, error) {
	params := map[string]int64{
		"requested_mb": requestedMB,
	}

	result, err := c.call("audit_resources", params)
	if err != nil {
		return false, 0, 0, false, err
	}

	var resp struct {
		Available bool  `json:"available"`
		TotalMB   int64 `json:"total_mb"`
		FreeMB    int64 `json:"free_mb"`
		Allowed   bool  `json:"allowed"`
	}
	if err := json.Unmarshal(result, &resp); err != nil {
		return false, 0, 0, false, fmt.Errorf("parse audit response: %w", err)
	}

	return resp.Available, resp.TotalMB, resp.FreeMB, resp.Allowed, nil
}

// ValidatePrompt validates a prompt via cograw.
func (c *Client) ValidatePrompt(prompt string, routingHints map[string][]string) (string, string, string, string, error) {
	params := map[string]interface{}{
		"prompt": prompt,
	}
	if len(routingHints) > 0 {
		params["routing_hints"] = routingHints
	}

	result, err := c.call("validate_prompt", params)
	if err != nil {
		return "", "", "", "", err
	}

	var resp struct {
		Action         string `json:"action"`
		ModifiedPrompt string `json:"modified_prompt,omitempty"`
		Reason         string `json:"reason,omitempty"`
		Model          string `json:"model,omitempty"`
	}
	if err := json.Unmarshal(result, &resp); err != nil {
		return "", "", "", "", fmt.Errorf("parse validate_prompt response: %w", err)
	}

	return resp.Action, resp.ModifiedPrompt, resp.Reason, resp.Model, nil
}

// ValidatePackageRequest validates a package request via cograw.
func (c *Client) ValidatePackageRequest(params ValidationParams) (ValidationResult, error) {
	result, err := c.call("validate_package_request", params)
	if err != nil {
		return ValidationResult{}, err
	}

	var resp ValidationResult
	if err := json.Unmarshal(result, &resp); err != nil {
		return ValidationResult{}, fmt.Errorf("parse validate_package_request response: %w", err)
	}

	return resp, nil
}

// IsReady returns true if the client is connected and cograw has responded to healthcheck.
func (c *Client) IsReady() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil
}

// Close closes the connection to cograw.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

// call sends a JSON-RPC 2.0 request and reads the matching response.
func (c *Client) call(method string, params interface{}) (json.RawMessage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil, fmt.Errorf("cograw not connected")
	}

	c.id++

	var rawParams json.RawMessage
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("marshal params: %w", err)
		}
		rawParams = b
	}

	req := RPCRequest{
		JSONRPC: "2.0",
		ID:      c.id,
		Method:  method,
		Params:  rawParams,
	}

	encoder := json.NewEncoder(c.conn)
	if err := encoder.Encode(req); err != nil {
		c.conn.Close()
		c.conn = nil
		return nil, fmt.Errorf("send to cograw: %w", err)
	}

	var resp RPCResponse
	decoder := json.NewDecoder(c.conn)
	if err := decoder.Decode(&resp); err != nil {
		c.conn.Close()
		c.conn = nil
		return nil, fmt.Errorf("read from cograw: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("%s: %s", resp.Error.Code, resp.Error.Message)
	}

	if resp.Result == nil {
		return nil, fmt.Errorf("no result from cograw")
	}

	return resp.Result, nil
}
