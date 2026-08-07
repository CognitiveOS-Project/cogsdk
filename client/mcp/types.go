/*
Package mcp provides MCP JSON-RPC 2.0 framing, tool-call wrappers, and schema
validation per product-specs/specs/mcp-conventions.md.

A Server runs the stdio transport side (bridges announce mcp.list_tools and
answer tool calls), and a Client drives the other side (the daemon spawning a
bridge, or any tool caller). Tool definitions carry an inputSchema that can be
validated against JSON Schema before dispatch.
*/
package mcp

import "encoding/json"

// JSON-RPC protocol version used by all CognitiveOS MCP servers.
const JSONRPCVersion = "2.0"

// rpcRequest is a JSON-RPC 2.0 request or notification.
type rpcRequest struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params,omitempty"`
}

// rpcResponse is a JSON-RPC 2.0 response.
type rpcResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *rpcError        `json:"error,omitempty"`
}

// rpcError is the JSON-RPC error object.
type rpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ContentItem is one piece of tool output.
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Result is the MCP tool-call result envelope (mcp-conventions.md).
type Result struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// Text returns the concatenated text of all content items.
func (r Result) Text() string {
	var s string
	for _, c := range r.Content {
		s += c.Text
	}
	return s
}
