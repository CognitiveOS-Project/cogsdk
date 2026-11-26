package prompt

import (
	"strings"
	"testing"
)

func TestNewPrompt(t *testing.T) {
	p := NewPrompt()
	if p == nil {
		t.Fatal("expected non-nil prompt")
	}
	if p.SystemPrompt() != "" {
		t.Fatalf("expected empty system prompt, got %q", p.SystemPrompt())
	}
	if len(p.Messages()) != 0 {
		t.Fatalf("expected empty messages, got %v", p.Messages())
	}
	if len(p.Context()) != 0 {
		t.Fatalf("expected empty context, got %v", p.Context())
	}
}

func TestPromptWithSystemPrompt(t *testing.T) {
	p := NewPrompt().WithSystemPrompt("You are a helpful assistant.")
	if p.SystemPrompt() != "You are a helpful assistant." {
		t.Fatalf("expected system prompt, got %q", p.SystemPrompt())
	}
}

func TestPromptWithUserMessage(t *testing.T) {
	p := NewPrompt()
	p.WithUserMessage("hello")
	p.WithUserMessage("world")
	msgs := p.Messages()
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0] != "hello" || msgs[1] != "world" {
		t.Fatalf("unexpected messages: %v", msgs)
	}
}

func TestPromptWithContext(t *testing.T) {
	p := NewPrompt()
	p.WithContext("key1", "value1")
	p.WithContext("key2", "value2")
	ctx := p.Context()
	if len(ctx) != 2 {
		t.Fatalf("expected 2 context entries, got %d", len(ctx))
	}
	if ctx["key1"] != "value1" || ctx["key2"] != "value2" {
		t.Fatalf("unexpected context: %v", ctx)
	}

	// Remove a key
	p.WithContext("key1", "")
	ctx = p.Context()
	if _, ok := ctx["key1"]; ok {
		t.Fatal("expected key1 to be removed")
	}
	if len(ctx) != 1 {
		t.Fatalf("expected 1 context entry, got %d", len(ctx))
	}
}

func TestPromptBuild(t *testing.T) {
	p := NewPrompt().
		WithSystemPrompt("You are a helpful assistant.").
		WithContext("user", "alice").
		WithContext("lang", "en").
		WithUserMessage("hello").
		WithUserMessage("how are you")

	out := p.Build()
	if !strings.Contains(out, "SYSTEM: You are a helpful assistant.") {
		t.Fatalf("expected SYSTEM line, got:\n%s", out)
	}
	if !strings.Contains(out, "CONTEXT:") {
		t.Fatalf("expected CONTEXT section, got:\n%s", out)
	}
	if !strings.Contains(out, "USER:") {
		t.Fatalf("expected USER section, got:\n%s", out)
	}
	if !strings.Contains(out, "hello") {
		t.Fatalf("expected user message 'hello', got:\n%s", out)
	}
	if !strings.Contains(out, "how are you") {
		t.Fatalf("expected user message 'how are you', got:\n%s", out)
	}

	// Context keys should be sorted
	ctxIdx := strings.Index(out, "CONTEXT:")
	afterCtx := out[ctxIdx:]
	if strings.Index(afterCtx, "lang") > strings.Index(afterCtx, "user") {
		t.Fatalf("expected context keys sorted, got:\n%s", afterCtx)
	}
}

func TestPromptBuildNoSystem(t *testing.T) {
	p := NewPrompt().WithUserMessage("hello")
	out := p.Build()
	if strings.Contains(out, "SYSTEM:") {
		t.Fatalf("expected no SYSTEM line, got:\n%s", out)
	}
	if !strings.Contains(out, "USER:") {
		t.Fatalf("expected USER section, got:\n%s", out)
	}
}

func TestPromptBuildNoContext(t *testing.T) {
	p := NewPrompt().WithSystemPrompt("sys").WithUserMessage("hello")
	out := p.Build()
	if strings.Contains(out, "CONTEXT:") {
		t.Fatalf("expected no CONTEXT section, got:\n%s", out)
	}
}

func TestPromptValidate(t *testing.T) {
	p := NewPrompt().WithUserMessage("hello")
	if err := p.Validate(); err != nil {
		t.Fatalf("expected valid prompt, got error: %v", err)
	}

	p = NewPrompt()
	err := p.Validate()
	if err == nil {
		t.Fatal("expected error for empty prompt")
	}
	if err.Error() != "prompt must contain at least one user message" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestPromptString(t *testing.T) {
	p := NewPrompt().WithUserMessage("hello")
	if p.String() != p.Build() {
		t.Fatal("String() should match Build()")
	}
}

func TestPromptMarshalJSON(t *testing.T) {
	p := NewPrompt().
		WithSystemPrompt("sys").
		WithUserMessage("hello")
	data, err := p.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	if string(data) == "" {
		t.Fatal("expected non-empty JSON")
	}
}
