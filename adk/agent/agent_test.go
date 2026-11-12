package agent

import (
	"testing"
)

func TestAgentCreation(t *testing.T) {
	a := NewAgent("test-agent")
	if a.ID() == "" {
		t.Fatal("expected non-empty ID")
	}
	if a.Name() != "test-agent" {
		t.Fatalf("expected name test-agent, got %s", a.Name())
	}
	if a.State() != StateIdle {
		t.Fatalf("expected initial state idle, got %s", a.State())
	}
	if len(a.Goals()) != 0 {
		t.Fatalf("expected empty goals, got %v", a.Goals())
	}
	if len(a.Tools()) != 0 {
		t.Fatalf("expected empty tools, got %v", a.Tools())
	}
}

func TestAgentSetGoals(t *testing.T) {
	a := NewAgent("test-agent")
	a.SetGoals([]string{"goal-1", "goal-2"})
	if a.State() != StateListening {
		t.Fatalf("expected listening state after setting goals, got %s", a.State())
	}
	goals := a.Goals()
	if len(goals) != 2 {
		t.Fatalf("expected 2 goals, got %d", len(goals))
	}
	if goals[0] != "goal-1" || goals[1] != "goal-2" {
		t.Fatalf("expected goals [goal-1 goal-2], got %v", goals)
	}
}

func TestAgentRegisterTool(t *testing.T) {
	a := NewAgent("test-agent")
	a.RegisterTool("cognitiveos.display.render_image")
	a.RegisterTool("cognitiveos.audio.play")
	tools := a.Tools()
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
	if tools[0] != "cognitiveos.display.render_image" || tools[1] != "cognitiveos.audio.play" {
		t.Fatalf("unexpected tools: %v", tools)
	}

	// Duplicate registration should not add duplicate
	a.RegisterTool("cognitiveos.display.render_image")
	tools = a.Tools()
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools after duplicate, got %d", len(tools))
	}
}

func TestAgentSetCapability(t *testing.T) {
	a := NewAgent("test-agent")
	a.SetCapability("vision", "enabled")
	cap := a.Capability()
	if cap["vision"] != "enabled" {
		t.Fatalf("expected capability vision=enabled, got %v", cap["vision"])
	}
}

func TestAgentMarshalJSON(t *testing.T) {
	a := NewAgent("json-agent")
	a.SetGoals([]string{"g1"})
	a.RegisterTool("tool-1")
	data, err := a.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	if string(data) == "" {
		t.Fatal("expected non-empty JSON")
	}
}