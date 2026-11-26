// Package coginfer provides a client for the coginfer Wide Model inference server.
//
// coginfer is a Wide Model HTTP inference server that exposes an Ollama-compatible
// API. The CoginferClient provides methods to load, generate, unload, and health
// check the Wide Model via HTTP.
package coginfer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// DefaultEndpoint returns the coginfer HTTP endpoint.
func DefaultEndpoint() string {
	return "http://127.0.0.1:11434"
}

// Client is an HTTP client for the coginfer Wide Model inference server.
//
// CoginferClient communicates over HTTP using an Ollama-compatible API.
// All methods are safe for concurrent use.
type Client struct {
	endpoint    string
	httpClient  *http.Client
	loaded      bool
	modelName   string
	modelID     string
	systemPrompt string
	mu          sync.RWMutex
}

// NewClient creates a new CoginferClient connected to the given endpoint.
// If endpoint is empty, DefaultEndpoint() is used.
func NewClient(endpoint string) *Client {
	if endpoint == "" {
		endpoint = DefaultEndpoint()
	}
	return &Client{
		endpoint:   endpoint,
		httpClient: &http.Client{Timeout: 300 * time.Second},
	}
}

// Dial connects to the default coginfer endpoint.
func Dial() (*Client, error) {
	return NewClient(DefaultEndpoint()), nil
}

// DialPath connects to an explicit endpoint.
func DialPath(endpoint string) (*Client, error) {
	return NewClient(endpoint), nil
}

// Generate sends a prompt to coginfer and returns the generated response.
func (c *Client) Generate(prompt string) (string, error) {
	c.mu.RLock()
	sysPrompt := c.systemPrompt
	c.mu.RUnlock()

	fullPrompt := prompt
	if sysPrompt != "" {
		fullPrompt = sysPrompt + "\n\n" + prompt
	}

	body := map[string]interface{}{
		"model":  "wide-model",
		"prompt": fullPrompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature":  0.7,
			"num_predict":  512,
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return "", fmt.Errorf("encode request: %w", err)
	}

	resp, err := c.httpClient.Post(c.endpoint+"/api/generate", "application/json", &buf)
	if err != nil {
		return "", fmt.Errorf("inference request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("inference error (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	c.mu.Lock()
	c.loaded = true
	c.modelName = "wide-model"
	c.mu.Unlock()

	return result.Response, nil
}

// SetSystemPrompt sets the system prompt for subsequent generates.
func (c *Client) SetSystemPrompt(prompt string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.systemPrompt = prompt
}

// SystemPrompt returns the current system prompt.
func (c *Client) SystemPrompt() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.systemPrompt
}

// Load loads a model from the given path.
func (c *Client) Load(modelPath string) error {
	return c.LoadWithID(modelPath, "")
}

// LoadWithID loads a model from the given path with an optional model ID.
func (c *Client) LoadWithID(modelPath, modelID string) error {
	body := map[string]interface{}{
		"model": "wide-model",
		"path":  modelPath,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return fmt.Errorf("encode load request: %w", err)
	}

	resp, err := c.httpClient.Post(c.endpoint+"/api/pull", "application/json", &buf)
	if err != nil {
		return fmt.Errorf("load request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("load failed: HTTP %d", resp.StatusCode)
	}

	c.mu.Lock()
	c.loaded = true
	c.modelName = modelPath
	c.modelID = modelID
	c.mu.Unlock()

	return nil
}

// Unload unloads the current model.
func (c *Client) Unload(reason string) error {
	body := map[string]interface{}{
		"reason": reason,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return fmt.Errorf("encode unload request: %w", err)
	}

	req, err := http.NewRequest("DELETE", c.endpoint+"/api/delete", &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unload request: %w", err)
	}
	defer resp.Body.Close()

	c.mu.Lock()
	c.loaded = false
	c.modelName = ""
	c.modelID = ""
	c.mu.Unlock()

	return nil
}

// LoadAdapter loads a model adapter.
func (c *Client) LoadAdapter(adapterPath string) error {
	body := map[string]interface{}{
		"adapter": adapterPath,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return fmt.Errorf("encode adapter request: %w", err)
	}

	resp, err := c.httpClient.Post(c.endpoint+"/api/adapter", "application/json", &buf)
	if err != nil {
		return fmt.Errorf("adapter load request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("adapter load failed: HTTP %d", resp.StatusCode)
	}

	return nil
}

// IsLoaded returns true if a model is currently loaded.
func (c *Client) IsLoaded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loaded
}

// LoadedModelName returns the name of the currently loaded model.
func (c *Client) LoadedModelName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.loaded {
		return ""
	}
	return c.modelName
}

// LoadedModelID returns the ID of the currently loaded model.
func (c *Client) LoadedModelID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if !c.loaded {
		return ""
	}
	return c.modelID
}

// Health checks if coginfer is healthy.
func (c *Client) Health() bool {
	resp, err := c.httpClient.Get(c.endpoint + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}
