package daemon

// Message types exchanged over the cognitiveosd Unix socket.
const (
	MsgInputForward    = "input_forward"
	MsgOutputDeliver   = "output_deliver"
	MsgSystemCode      = "system_code"
	MsgMCPRegister     = "mcp_register"
	MsgMCPUnregister   = "mcp_unregister"
	MsgMCPResult       = "mcp_result"
	MsgAuditRequest    = "audit_request"
	MsgAuditReport     = "audit_report"
	MsgStatusRequest   = "status_request"
	MsgWideModelLoad   = "wide_model_load"
	MsgWideModelUnload = "wide_model_unload"
	MsgShutdownNotice  = "shutdown_notice"
)

// InputPayload is the payload of input_forward.
type InputPayload struct {
	Mode    string        `json:"mode"`
	Content string        `json:"content"`
	Context *InputContext `json:"context,omitempty"`
}

// InputContext carries session and device context for input_forward.
type InputContext struct {
	SessionID string `json:"session_id"`
	Device    string `json:"device,omitempty"`
}

// OutputPayload is the payload of output_deliver.
type OutputPayload struct {
	SessionID   string     `json:"session_id"`
	Content     string     `json:"content"`
	ContentType string     `json:"content_type"`
	Media       *MediaInfo `json:"media,omitempty"`
	Actions     []Action   `json:"actions,omitempty"`
}

// MediaInfo describes media attached to an output delivery.
type MediaInfo struct {
	Type          string   `json:"type"`
	Paths         []string `json:"paths"`
	RenderCommand string   `json:"render_command"`
}

// Action is a suggested user-facing action attached to output.
type Action struct {
	Label   string `json:"label"`
	Command string `json:"command"`
}

// SystemCodePayload is the payload of system_code.
type SystemCodePayload struct {
	Code       string `json:"code"`
	UnlockCode string `json:"unlock_code,omitempty"`
	Origin     string `json:"origin"`
}

// CodeAcceptedPayload is the payload of a code_accepted response.
type CodeAcceptedPayload struct {
	Status string `json:"status"`
	Effect string `json:"effect"`
}

// MCPRegisterPayload is the payload of mcp_register.
type MCPRegisterPayload struct {
	Server MCPInfo   `json:"server"`
	Tools  []MCPTool `json:"tools"`
}

// MCPInfo identifies a registering MCP server.
type MCPInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Transport string `json:"transport"`
	PID       int    `json:"pid,omitempty"`
}

// MCPTool describes a single MCP tool capability.
type MCPTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
}

// MCPRegisteredPayload is the payload of an mcp_registered response.
type MCPRegisteredPayload struct {
	Status          string   `json:"status"`
	ServerID        string   `json:"server_id"`
	RegisteredTools []string `json:"registered_tools"`
}

// MCPUnregisterPayload is the payload of mcp_unregister.
type MCPUnregisterPayload struct {
	Reason string `json:"reason"`
}

// ToolCall is a parsed tool invocation from the Wide Model.
type ToolCall struct {
	Tool      string                 `json:"tool"`
	Arguments map[string]interface{} `json:"arguments"`
}

// MCPResultPayload is the payload of mcp_result.
type MCPResultPayload struct {
	Status  string        `json:"status"`
	Error   *ErrorInfo    `json:"error,omitempty"`
	Content []ContentItem `json:"content,omitempty"`
}

// ContentItem is one text content item of an MCP result.
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// StatusRequestPayload is the (empty) payload of status_request.
type StatusRequestPayload struct{}

// StatusResponsePayload is the payload of a status_response.
type StatusResponsePayload struct {
	State            string          `json:"state"`
	UptimeSeconds    int64           `json:"uptime_seconds"`
	WideModel        WideModelStatus `json:"wide_model"`
	ModelRegistry    int             `json:"model_registry_count"`
	PatchesInstalled int             `json:"patches_installed"`
	MCPServersActive int             `json:"mcp_servers_active"`
}

// WideModelStatus describes the loaded Wide Model.
type WideModelStatus struct {
	Status  string `json:"status"`
	Name    string `json:"name,omitempty"`
	ModelID string `json:"model_id,omitempty"`
}

// AuditRequestPayload is the (empty) payload of audit_request.
type AuditRequestPayload struct{}

// AuditReportPayload is the payload of an audit_report response.
type AuditReportPayload struct {
	Timestamp string         `json:"timestamp"`
	Resources AuditResources `json:"resources"`
}

// AuditResources is a hardware resource snapshot.
type AuditResources struct {
	RAM     RAMInfo     `json:"ram"`
	Storage StorageInfo `json:"storage"`
	CPU     CPUInfo     `json:"cpu"`
	NPU     NPUInfo     `json:"npu"`
	Network NetworkInfo `json:"network"`
}

// RAMInfo describes system memory.
type RAMInfo struct {
	TotalMB     int64 `json:"total_mb"`
	AvailableMB int64 `json:"available_mb"`
	UsedByAIMB  int64 `json:"used_by_ai_mb"`
}

// StorageInfo describes storage.
type StorageInfo struct {
	TotalMB     int64 `json:"total_mb"`
	AvailableMB int64 `json:"available_mb"`
	PatchesMB   int64 `json:"patches_mb"`
	ModelsMB    int64 `json:"models_mb"`
}

// CPUInfo describes the CPU.
type CPUInfo struct {
	Cores       int     `json:"cores"`
	LoadPercent float64 `json:"load_percent"`
}

// NPUInfo describes an available NPU.
type NPUInfo struct {
	Available bool   `json:"available"`
	Model     string `json:"model,omitempty"`
	MemoryMB  int64  `json:"memory_mb,omitempty"`
}

// NetworkInfo describes network connectivity.
type NetworkInfo struct {
	Connected     bool   `json:"connected"`
	Interface     string `json:"interface,omitempty"`
	SignalPercent int    `json:"signal_percent,omitempty"`
}

// ShutdownNoticePayload is the payload of shutdown_notice.
type ShutdownNoticePayload struct {
	Reason string `json:"reason"`
}
