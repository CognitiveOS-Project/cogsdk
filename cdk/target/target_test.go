package target

import (
	"testing"
)

func TestProvisioner(t *testing.T) {
	p := NewProvisioner()

	// Register targets
	p.RegisterTarget("target-1", "http://target-1.example.com", "us-east-1")
	p.RegisterTarget("target-2", "http://target-2.example.com", "us-west-2")

	// Verify targets are registered
	t1, err := p.GetTarget("target-1")
	if err != nil {
		t.Fatalf("GetTarget failed: %v", err)
	}
	if t1.Name != "target-1" || t1.Status != "pending" {
		t.Fatalf("expected target-1 pending, got %+v", t1)
	}

	t2, err := p.GetTarget("target-2")
	if err != nil {
		t.Fatalf("GetTarget failed: %v", err)
	}
	if t2.Name != "target-2" {
		t.Fatalf("expected target-2, got %s", t2.Name)
	}

	// List all targets
	targets := p.ListTargets()
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}

	// Provision target-1
	if err := p.Provision("target-1"); err != nil {
		t.Fatalf("Provision failed: %v", err)
	}

	t1, _ = p.GetTarget("target-1")
	if t1.Status != "active" {
		t.Fatalf("expected active status, got %s", t1.Status)
	}

	// Provision non-existent target should fail
	if err := p.Provision("nonexistent"); err == nil {
		t.Fatal("expected error for nonexistent target")
	}
}