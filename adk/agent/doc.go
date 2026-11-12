// Package agent provides the agent toolkit: agent lifecycle, state management,
// tool orchestration, and capability discovery for CognitiveOS agents.
//
// An Agent is a long-running, stateful process that observes inputs, reasons
// about goals, and executes actions through MCP tool calls. Agent state is
// persisted to allow resumption across restarts, and all agent actions are
// mediated through the cpm daemon per ADR-004 (validated tool domains).
package agent