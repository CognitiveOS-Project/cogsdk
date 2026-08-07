package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
)

// Client drives an MCP server over an io.ReadWriter pair (e.g. the stdin/stdout
// of a spawned bridge, or a net.Pipe in tests). It can list tools and invoke
// tool calls.
type Client struct {
	enc     *json.Encoder
	scanner *bufio.Scanner
	mu      sync.Mutex
	id      int64
	nextID  func() string
}

// NewClient creates an MCP client writing requests to w and reading responses
// from r.
func NewClient(r io.Reader, w io.Writer) *Client {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	return &Client{
		enc:     json.NewEncoder(w),
		scanner: scanner,
	}
}

// SetIDFunc overrides the request-ID generator (default: an incrementing counter).
func (c *Client) SetIDFunc(f func() string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nextID = f
}

func (c *Client) newID() string {
	if c.nextID != nil {
		return c.nextID()
	}
	c.id++
	return fmt.Sprintf("%d", c.id)
}

// ListTools fetches the server's tool list via mcp.list_tools.
func (c *Client) ListTools() ([]Tool, error) {
	req := rpcRequest{
		JSONRPC: JSONRPCVersion,
		ID:      rawID(c.newID()),
		Method:  "mcp.list_tools",
	}

	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected list_tools result")
	}

	toolsRaw, _ := result["tools"].([]interface{})
	tools := make([]Tool, 0, len(toolsRaw))
	for _, t := range toolsRaw {
		data, err := json.Marshal(t)
		if err != nil {
			continue
		}
		var tool Tool
		if err := json.Unmarshal(data, &tool); err != nil {
			continue
		}
		tools = append(tools, tool)
	}
	return tools, nil
}

// Call invokes a tool with the given arguments and returns the MCP result.
func (c *Client) Call(tool string, args map[string]interface{}) (Result, error) {
	if args == nil {
		args = make(map[string]interface{})
	}
	params, err := json.Marshal(map[string]interface{}{"arguments": args})
	if err != nil {
		return Result{}, fmt.Errorf("marshal params: %w", err)
	}

	req := rpcRequest{
		JSONRPC: JSONRPCVersion,
		ID:      rawID(c.newID()),
		Method:  tool,
		Params:  params,
	}

	resp, err := c.roundTrip(req)
	if err != nil {
		return Result{}, err
	}

	if resp.Error != nil {
		return Result{}, fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		return Result{}, fmt.Errorf("unexpected tool result")
	}

	out := Result{}
	if isErr, _ := result["isError"].(bool); isErr {
		out.IsError = true
	}
	if content, ok := result["content"].([]interface{}); ok {
		for _, c := range content {
			if cm, ok := c.(map[string]interface{}); ok {
				out.Content = append(out.Content, ContentItem{
					Type: fmt.Sprintf("%v", cm["type"]),
					Text: fmt.Sprintf("%v", cm["text"]),
				})
			}
		}
	}
	return out, nil
}

func (c *Client) roundTrip(req rpcRequest) (*rpcResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.enc.Encode(req); err != nil {
		return nil, fmt.Errorf("send to MCP server: %w", err)
	}

	for c.scanner.Scan() {
		var resp rpcResponse
		if err := json.Unmarshal(c.scanner.Bytes(), &resp); err != nil {
			return nil, fmt.Errorf("parse MCP response: %w", err)
		}
		return &resp, nil
	}
	if err := c.scanner.Err(); err != nil {
		return nil, fmt.Errorf("read MCP response: %w", err)
	}
	return nil, fmt.Errorf("MCP server disconnected")
}

// rawID builds a *json.RawMessage from an id string.
func rawID(id string) *json.RawMessage {
	raw := json.RawMessage(fmt.Sprintf("%q", id))
	return &raw
}

// ParseErrorText splits "ERROR:<code>:<message>" into its components.
func ParseErrorText(text string) (code, message string) {
	text = strings.TrimPrefix(text, "ERROR:")
	parts := strings.SplitN(text, ":", 2)
	code = parts[0]
	if len(parts) == 2 {
		message = strings.TrimSpace(parts[1])
	}
	return code, message
}
