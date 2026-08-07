package env

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SecretsFileName is the file name used for the secrets store.
const SecretsFileName = "secrets.json"

// Secret scopes.
const (
	ScopePatch  = "patch"
	ScopeGlobal = "global"
)

// Secrets holds key-value secret pairs.
type Secrets struct {
	Secrets map[string]string `json:"secrets"`
}

// NewSecrets creates an empty Secrets instance.
func NewSecrets() *Secrets {
	return &Secrets{Secrets: make(map[string]string)}
}

// SecretsConfig overrides the default secret store locations.
type SecretsConfig struct {
	// PatchBase is the directory containing <patch>/secrets.json files.
	// Default: /cognitiveos/lib/cpm/secrets
	PatchBase string
	// GlobalDir is the directory containing the global secrets.json.
	// Default: $HOME/.cpm
	GlobalDir string
}

// SecretsStore loads and saves secrets in patch/global scopes.
type SecretsStore struct {
	cfg SecretsConfig
}

// DefaultSecretsStore returns a store using the default locations.
func DefaultSecretsStore() *SecretsStore {
	return &SecretsStore{}
}

// NewSecretsStore returns a store with explicit locations.
func NewSecretsStore(cfg SecretsConfig) *SecretsStore {
	return &SecretsStore{cfg: cfg}
}

// SecretsPath returns the secrets.json path for the given scope.
func (s *SecretsStore) SecretsPath(scope, patchName string) (string, error) {
	switch scope {
	case ScopePatch:
		if patchName == "" {
			return "", fmt.Errorf("patch name required for patch scope")
		}
		base := s.cfg.PatchBase
		if base == "" {
			if d := os.Getenv("CPM_PATCHES_DIR"); d != "" {
				base = filepath.Join(filepath.Dir(d), "lib", "cpm", "secrets")
			} else {
				base = "/cognitiveos/lib/cpm/secrets"
			}
		}
		return filepath.Join(base, patchName, SecretsFileName), nil
	case ScopeGlobal:
		dir := s.cfg.GlobalDir
		if dir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("cannot determine home directory: %w", err)
			}
			dir = filepath.Join(home, ".cpm")
		}
		return filepath.Join(dir, SecretsFileName), nil
	default:
		return "", fmt.Errorf("unknown scope %q (use 'patch' or 'global')", scope)
	}
}

// Load reads secrets for the given scope. A missing file yields an empty store.
func (s *SecretsStore) Load(scope, patchName string) (*Secrets, error) {
	path, err := s.SecretsPath(scope, patchName)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewSecrets(), nil
		}
		return nil, fmt.Errorf("read secrets: %w", err)
	}

	var out Secrets
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parse secrets: %w", err)
	}
	if out.Secrets == nil {
		out.Secrets = make(map[string]string)
	}
	return &out, nil
}

// Save writes secrets for the given scope with 0600 file / 0700 dir perms.
func (s *SecretsStore) Save(scope, patchName string, secrets *Secrets) error {
	path, err := s.SecretsPath(scope, patchName)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create secrets directory: %w", err)
	}

	data, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal secrets: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write secrets: %w", err)
	}
	return nil
}

// Delete removes the secrets file for the given scope.
func (s *SecretsStore) Delete(scope, patchName string) error {
	path, err := s.SecretsPath(scope, patchName)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove secrets: %w", err)
	}
	return nil
}

// Merge returns a new Secrets with global values overridden by patch values.
func Merge(global, patch *Secrets) *Secrets {
	merged := NewSecrets()
	if global != nil {
		for k, v := range global.Secrets {
			merged.Secrets[k] = v
		}
	}
	if patch != nil {
		for k, v := range patch.Secrets {
			merged.Secrets[k] = v
		}
	}
	return merged
}

// SecretsLookup builds a Lookup from a merged Secrets store.
func SecretsLookup(s *Secrets) Lookup {
	if s == nil {
		return func(string) (string, bool) { return "", false }
	}
	return func(name string) (string, bool) {
		v, ok := s.Secrets[name]
		return v, ok
	}
}
