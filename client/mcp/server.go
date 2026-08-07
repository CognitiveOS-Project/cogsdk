package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// Tool describes an MCP tool capability.
type Tool struct {
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	InputSchema  interface{} `json:"inputSchema"`
	OutputSchema interface{} `json:"outputSchema,omitempty"`
	Annotations  interface{} `json:"annotations,omitempty"`
}

// Handler executes a tool call with its arguments.
type Handler func(map[string]interface{}) (interface{}, error)

// Server is a stdio-transport MCP server (mcp-conventions.md: Transport —
// Stdio is Preferred). It reads JSON-RPC 2.0 lines from stdin, answers
// mcp.list_tools, dispatches tool-call methods, and responds to the
// "healthcheck" notification.
type Server struct {
	Name     string
	Version  string
	Tools    []Tool
	handlers map[string]Handler
	started  time.Time
	logger   *log.Logger
	mu       sync.Mutex
}

// New creates an MCP server with the given name.
func New(name string) *Server {
	return NewWithVersion(name, "0.1.0")
}

// NewWithVersion creates an MCP server with an explicit version.
func NewWithVersion(name, version string) *Server {
	return &Server{
		Name:     name,
		Version:  version,
		handlers: make(map[string]Handler),
		started:  time.Now(),
		logger:   log.New(os.Stderr, "", 0),
	}
}

// SetLogger overrides the default stderr logger.
func (s *Server) SetLogger(l *log.Logger) {
	s.logger = l
}

// Handle registers a handler for a tool name.
func (s *Server) Handle(tool string, fn Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[tool] = fn
}

// Run blocks serving requests from the provided reader, writing to the
// provided writer. Typical usage: Run(os.Stdin, os.Stdout).
func (s *Server) Run(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		s.handleMessage(scanner.Text(), w)
	}
	return scanner.Err()
}

// RunStdio runs the server over os.Stdin/os.Stdout.
func (s *Server) RunStdio() error {
	return s.Run(os.Stdin, os.Stdout)
}

func (s *Server) handleMessage(line string, w io.Writer) {
	var req rpcRequest
	if err := json.Unmarshal([]byte(line), &req); err != nil {
		return
	}
	if req.ID == nil {
		s.handleNotification(req.Method, req.Params, w)
		return
	}
	resp := s.dispatch(&req)
	reply, _ := json.Marshal(resp)
	_, _ = fmt.Fprintln(w, string(reply))
}

func (s *Server) handleNotification(method string, params json.RawMessage, w io.Writer) {
	switch method {
	case "healthcheck":
		resp := map[string]interface{}{
			"type":           "healthcheck_ok",
			"uptime_seconds": int(time.Since(s.started).Seconds()),
			"tools_healthy":  true,
		}
		reply, _ := json.Marshal(resp)
		_, _ = fmt.Fprintln(w, string(reply))
	}
}

func (s *Server) dispatch(req *rpcRequest) *rpcResponse {
	switch req.Method {
	case "mcp.list_tools":
		return s.handleListTools(req.ID)
	default:
		return s.handleToolCall(req.ID, req.Method, req.Params)
	}
}

func (s *Server) handleListTools(id *json.RawMessage) *rpcResponse {
	s.mu.Lock()
	tools := make([]Tool, len(s.Tools))
	copy(tools, s.Tools)
	s.mu.Unlock()

	return &rpcResponse{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

func (s *Server) handleToolCall(id *json.RawMessage, method string, params json.RawMessage) *rpcResponse {
	s.mu.Lock()
	handler, ok := s.handlers[method]
	s.mu.Unlock()

	if !ok {
		return &rpcResponse{
			JSONRPC: JSONRPCVersion,
			ID:      id,
			Result: Result{
				IsError: true,
				Content: []ContentItem{{Type: "text", Text: "ERROR:E_NOT_FOUND: tool not found: " + method}},
			},
		}
	}

	var args map[string]interface{}
	if len(params) > 0 {
		var callParams struct {
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(params, &callParams); err == nil {
			args = callParams.Arguments
		}
	}
	if args == nil {
		args = make(map[string]interface{})
	}

	result, err := handler(args)
	if err != nil {
		errMsg := err.Error()
		if !strings.HasPrefix(errMsg, "ERROR:") {
			if strings.HasPrefix(errMsg, "E_") {
				errMsg = "ERROR:" + errMsg
			} else {
				errMsg = "ERROR:E_INTERNAL: " + errMsg
			}
		}
		return &rpcResponse{
			JSONRPC: JSONRPCVersion,
			ID:      id,
			Result: Result{
				IsError: true,
				Content: []ContentItem{{Type: "text", Text: errMsg}},
			},
		}
	}

	text, ok := result.(string)
	if !ok {
		data, _ := json.Marshal(result)
		text = string(data)
	}

	return &rpcResponse{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Result: Result{
			Content: []ContentItem{{Type: "text", Text: text}},
		},
	}
}

// Log writes a formatted line at INFO level.
func (s *Server) Log(format string, args ...interface{}) {
	s.Logf("INFO", format, args...)
}

// Logf writes a formatted line at the given level.
func (s *Server) Logf(level, format string, args ...interface{}) {
	ts := time.Now().UTC().Format(time.RFC3339)
	msg := fmt.Sprintf(format, args...)
	s.logger.Printf("%s [%s] %s", ts, level, msg)
}
