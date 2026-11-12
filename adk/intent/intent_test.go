package intent

import (
	"testing"
)

func TestRecognizerEmpty(t *testing.T) {
	r := NewRecognizer()
	intent, err := r.Recognize("hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent != nil {
		t.Fatalf("expected nil intent for empty recognizer, got %v", intent)
	}
}

func TestRecognizerRegisterAndMatch(t *testing.T) {
	r := NewRecognizer()
	r.RegisterIntent("wake", []string{"wake up", "start"}, nil)
	r.RegisterIntent("idle", []string{"go to sleep", "shutdown"}, nil)

	intent, err := r.Recognize("wake up the system")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent == nil {
		t.Fatal("expected non-nil intent")
	}
	if intent.Name != "wake" {
		t.Fatalf("expected intent wake, got %s", intent.Name)
	}
}

func TestRecognizerNoMatch(t *testing.T) {
	r := NewRecognizer()
	r.RegisterIntent("wake", []string{"wake up"}, nil)

	intent, err := r.Recognize("unrelated input")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent != nil {
		t.Fatalf("expected nil intent for no match, got %v", intent)
	}
}

func TestIntentString(t *testing.T) {
	i := &Intent{Name: "test", Params: map[string]interface{}{"key": "val"}, Score: 0.85}
	s := i.String()
	if s != "test(score=0.85)" {
		t.Fatalf("expected 'test(score=0.85)', got %q", s)
	}
}