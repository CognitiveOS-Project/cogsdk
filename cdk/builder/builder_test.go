package builder

import (
	"os"
	"testing"
)

func TestNewBuilder(t *testing.T) {
	b := NewBuilder("")
	if b.WorkDir == "" {
		t.Fatal("expected non-empty WorkDir when not provided")
	}

	b2 := NewBuilder("/tmp/custom")
	if b2.WorkDir != "/tmp/custom" {
		t.Fatalf("expected /tmp/custom, got %s", b2.WorkDir)
	}
}

func TestBuildImage(t *testing.T) {
	b := NewBuilder(t.TempDir())

	// BuildImage should fail when no Dockerfile exists
	err := b.BuildImage("nonexistent.Dockerfile", "test-image")
	if err == nil {
		t.Fatal("expected error when Dockerfile does not exist")
	}
}

func TestBuilderWorkDir(t *testing.T) {
	tmpDir := t.TempDir()
	b := NewBuilder(tmpDir)
	if b.WorkDir != tmpDir {
		t.Fatalf("expected %s, got %s", tmpDir, b.WorkDir)
	}

	_ = os.RemoveAll(tmpDir)
}