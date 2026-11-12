package agent

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// State is the lifecycle state of an Agent.
type State string

const (
	StateIdle      State = "idle"
	StateListening State = "listening"
	StateThinking  State = "thinking"
	StateActing    State = "acting"
	StateError     State = "error"
	StateStopped   State = "stopped"
)

// Agent is a long-running process that observes inputs, reasons about goals,
// and executes actions through MCP tool calls.
//
// Agent state is persisted to allow resumption across restarts. All agent
// actions are mediated through the cpm daemon per ADR-004 (validated tool
// domains); cogsdk provides typed clients only — it never executes tools
// directly.
type Agent struct {
	mu     sync.RWMutex
	id     string
	name   string
	state  State
	goals  []string
	tools  []string
	cap map[string]interface{}
}

// NewAgent creates a new Agent with the given name.
//
// The agent starts in the Idle state with no goals or tools registered.
func NewAgent(name string) *Agent {
	return &Agent{
		id:    generateID(),
		name:  name,
		state: StateIdle,
	}
}

// ID returns the unique identifier of the agent.
func (a *Agent) ID() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.id
}

// Name returns the human-readable name of the agent.
func (a *Agent) Name() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.name
}

// State returns the current lifecycle state of the agent.
func (a *Agent) State() State {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

// Goals returns the current list of goals for the agent.
func (a *Agent) Goals() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	goals := make([]string, len(a.goals))
	copy(goals, a.goals)
	return goals
}

// Tools returns the list of tool names the agent is authorized to call.
func (a *Agent) Tools() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	tools := make([]string, len(a.tools))
	copy(tools, a.tools)
	return tools
}

// SetGoals replaces the agent's goal list.
func (a *Agent) SetGoals(goals []string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.goals = append([]string(nil), goals...)
	if len(a.goals) > 0 {
		a.state = StateListening
	}
}

// RegisterTool adds a tool name to the agent's authorized tool list.
func (a *Agent) RegisterTool(tool string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, t := range a.tools {
		if t == tool {
			return
		}
	}
	a.tools = append(a.tools, tool)
	if len(a.tools) > 0 && len(a.goals) > 0 {
		a.state = StateListening
	}
}

// Capability returns the capability map for the agent.
func (a *Agent) Capability() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()
	cap := make(map[string]interface{}, len(a.cap))
	for k, v := range a.cap {
		cap[k] = v
	}
	return cap
}

// SetCapability sets a single capability key/value pair.
func (a *Agent) SetCapability(key, value string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cap == nil {
		a.cap = make(map[string]interface{})
	}
	a.cap[key] = value
}

// MarshalJSON serializes the agent to JSON.
func (a *Agent) MarshalJSON() ([]byte, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return json.Marshal(map[string]interface{}{
		"id":    a.id,
		"name":  a.name,
		"state": string(a.state),
		"goals": a.goals,
		"tools": a.tools,
		"cap":   a.cap,
	})
}

// generateID creates a deterministic agent identifier.
func generateID() string {
	return fmt.Sprintf("agent-%d", time.Now().UnixNano())
}