// Package prompt provides prompt construction for the CognitiveOS Wide Model.
//
// Prompt builds structured prompts with system instructions, user messages,
// and context, suitable for submission to the Wide Model through the
// cogsdk client layer. All prompt construction is client-side; prompt
// validation and submission are mediated through cogsdk typed clients.
package prompt

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Prompt is a structured prompt for the CognitiveOS Wide Model.
//
// A Prompt consists of a system instruction, an ordered list of user
// messages, and optional context key/value pairs. The Build method renders
// the prompt into a single string suitable for submission to the Wide Model
// through a cogsdk client.
type Prompt struct {
	mu      sync.RWMutex
	system  string
	messages []string
	context map[string]string
}

// NewPrompt creates a new empty Prompt.
//
// The prompt starts with no system instruction, no messages, and no context.
// Use WithSystemPrompt, WithUserMessage, and WithContext to populate it.
func NewPrompt() *Prompt {
	return &Prompt{
		context: make(map[string]string),
	}
}

// SystemPrompt returns the current system instruction.
func (p *Prompt) SystemPrompt() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.system
}

// Messages returns a copy of the current user messages.
func (p *Prompt) Messages() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	msgs := make([]string, len(p.messages))
	copy(msgs, p.messages)
	return msgs
}

// Context returns a copy of the current context map.
func (p *Prompt) Context() map[string]string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	ctx := make(map[string]string, len(p.context))
	for k, v := range p.context {
		ctx[k] = v
	}
	return ctx
}

// WithSystemPrompt sets the system instruction for the prompt.
//
// The system instruction sets the behavior and constraints for the Wide
// Model. It is rendered before all user messages.
func (p *Prompt) WithSystemPrompt(system string) *Prompt {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.system = system
	return p
}

// WithUserMessage appends a user message to the prompt.
//
// User messages are rendered in order after the system instruction.
func (p *Prompt) WithUserMessage(message string) *Prompt {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages = append(p.messages, message)
	return p
}

// WithContext sets a context key/value pair.
//
// Context pairs are rendered as key-value lines after the system instruction
// and before user messages. Setting an empty value removes the key.
func (p *Prompt) WithContext(key, value string) *Prompt {
	p.mu.Lock()
	defer p.mu.Unlock()
	if value == "" {
		delete(p.context, key)
		return p
	}
	p.context[key] = value
	return p
}

// Build renders the prompt into a single string.
//
// The output format is:
//   SYSTEM: <system-instruction>
//   CONTEXT:
//     key: value
//   USER:
//     <message-1>
//     ...
//     <message-N>
//
// If no system instruction is set, the SYSTEM line is omitted. If no context
// is set, the CONTEXT section is omitted.
func (p *Prompt) Build() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var b strings.Builder

	if p.system != "" {
		b.WriteString("SYSTEM: ")
		b.WriteString(p.system)
		b.WriteString("\n")
	}

	if len(p.context) > 0 {
		b.WriteString("CONTEXT:\n")
		keys := make([]string, 0, len(p.context))
		for k := range p.context {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			b.WriteString("  ")
			b.WriteString(k)
			b.WriteString(": ")
			b.WriteString(p.context[k])
			b.WriteString("\n")
		}
	}

	if len(p.messages) > 0 {
		b.WriteString("USER:\n")
		for _, m := range p.messages {
			b.WriteString(m)
			b.WriteString("\n")
		}
	}

	return strings.TrimSpace(b.String())
}

// Validate checks that the prompt is well-formed.
//
// A valid prompt must have at least one user message. A system instruction
// and context are optional.
func (p *Prompt) Validate() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.messages) == 0 {
		return fmt.Errorf("prompt must contain at least one user message")
	}
	return nil
}

// MarshalJSON serializes the prompt to JSON.
func (p *Prompt) MarshalJSON() ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return json.Marshal(map[string]interface{}{
		"system":  p.system,
		"messages": p.messages,
		"context": p.context,
	})
}

// String returns the built prompt string.
func (p *Prompt) String() string {
	return p.Build()
}
