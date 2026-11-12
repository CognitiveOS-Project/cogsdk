package cloudinit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

// CloudInit generates cloud-init user-data configurations for CognitiveOS
// instances.
type CloudInit struct {
	// OutputDir is the directory where generated user-data files are written.
	OutputDir string
}

// NewCloudInit creates a new CloudInit generator.
func NewCloudInit(outputDir string) *CloudInit {
	return &CloudInit{OutputDir: outputDir}
}

// GenerateUserData renders a cloud-init user-data YAML from a template and
// writes it to the output directory.
func (c *CloudInit) GenerateUserData(name, templateStr string, data interface{}) (string, error) {
	tmpl, err := template.New(name).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render template: %w", err)
	}

	outPath := filepath.Join(c.OutputDir, name+".yaml")
	if err := os.WriteFile(outPath, buf.Bytes(), 0644); err != nil {
		return "", fmt.Errorf("write user-data: %w", err)
	}

	return outPath, nil
}

// GenerateJSON renders a JSON configuration and writes it to the output directory.
func (c *CloudInit) GenerateJSON(name string, data interface{}) (string, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal JSON: %w", err)
	}

	outPath := filepath.Join(c.OutputDir, name+".json")
	if err := os.WriteFile(outPath, b, 0644); err != nil {
		return "", fmt.Errorf("write JSON: %w", err)
	}

	return outPath, nil
}

// Now returns the current time formatted for cloud-init timestamps.
func Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}