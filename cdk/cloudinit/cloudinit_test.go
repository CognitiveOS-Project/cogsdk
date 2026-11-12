package cloudinit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCloudInitGenerateUserData(t *testing.T) {
	dir := t.TempDir()
	ci := NewCloudInit(dir)

	outPath, err := ci.GenerateUserData("test-config", "key: {{.Value}}", map[string]interface{}{"Value": "hello"})
	if err != nil {
		t.Fatalf("GenerateUserData failed: %v", err)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected output file to exist: %v", err)
	}

	if filepath.Base(outPath) != "test-config.yaml" {
		t.Fatalf("expected filename test-config.yaml, got %s", filepath.Base(outPath))
	}
}

func TestCloudInitGenerateJSON(t *testing.T) {
	dir := t.TempDir()
	ci := NewCloudInit(dir)

	type Config struct {
		Name string `json:"name"`
		Val  int    `json:"value"`
	}

	outPath, err := ci.GenerateJSON("test-config", Config{Name: "test", Val: 42})
	if err != nil {
		t.Fatalf("GenerateJSON failed: %v", err)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected output file to exist: %v", err)
	}

	if filepath.Base(outPath) != "test-config.json" {
		t.Fatalf("expected filename test-config.json, got %s", filepath.Base(outPath))
	}
}

func TestNow(t *testing.T) {
	now := Now()
	if now == "" {
		t.Fatal("expected non-empty timestamp")
	}
}