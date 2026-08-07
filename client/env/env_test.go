package env

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestResolveEnvironment(t *testing.T) {
	lk := func(name string) (string, bool) {
		switch name {
		case "NAME", "TAG":
			return "cogsdk", true
		}
		return "", false
	}

	out, err := Resolve([]byte("hello ${NAME} @${TAG}!"), lk)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if string(out) != "hello cogsdk @cogsdk!" {
		t.Fatalf("out = %q", out)
	}
}

func TestResolveMissing(t *testing.T) {
	out, err := Resolve([]byte("${NOPE}"), func(string) (string, bool) {
		return "", false
	})
	if err == nil {
		t.Fatalf("expected error, got %q", out)
	}
	if !errors.Is(err, ErrUndefined) {
		t.Fatalf("error = %v, want ErrUndefined", err)
	}
	var uerr *UndefinedError
	if !errors.As(err, &uerr) {
		t.Fatalf("cannot unwrap to UndefinedError: %v", err)
	}
	if uerr.Name != "NOPE" {
		t.Fatalf("name = %q", uerr.Name)
	}
}

func TestResolveNoPlaceholders(t *testing.T) {
	out, err := Resolve([]byte("no placeholders"), func(string) (string, bool) {
		t.Fatal("lookup should not be called")
		return "", false
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if string(out) != "no placeholders" {
		t.Fatalf("out = %q", out)
	}
}

func TestResolveOS(t *testing.T) {
	os.Setenv("COGSDK_TEST_VAR", "hi")
	defer os.Unsetenv("COGSDK_TEST_VAR")

	out, err := Resolve([]byte("${COGSDK_TEST_VAR}"), EnvLookup)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if string(out) != "hi" {
		t.Fatalf("out = %q", out)
	}
}

func TestSecretsStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := NewSecretsStore(SecretsConfig{PatchBase: dir, GlobalDir: dir})

	sec := NewSecrets()
	sec.Secrets["OPENAI_API_KEY"] = "sk-test"

	if err := s.Save(ScopePatch, "my-patch", sec); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := s.Load(ScopePatch, "my-patch")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Secrets["OPENAI_API_KEY"] != "sk-test" {
		t.Fatalf("loaded = %+v", got.Secrets)
	}

	path, err := s.SecretsPath(ScopePatch, "my-patch")
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("perm = %v, want 0600", info.Mode().Perm())
	}
}

func TestSecretsStoreMissingScope(t *testing.T) {
	s := NewSecretsStore(SecretsConfig{PatchBase: t.TempDir(), GlobalDir: t.TempDir()})
	if _, err := s.SecretsPath("bogus", ""); err == nil {
		t.Fatal("expected error for unknown scope")
	}
}

func TestSecretsMerge(t *testing.T) {
	g := NewSecrets()
	g.Secrets["A"] = "global-a"
	g.Secrets["B"] = "global-b"

	p := NewSecrets()
	p.Secrets["B"] = "patch-b"
	p.Secrets["C"] = "patch-c"

	merged := Merge(g, p)
	want := map[string]string{"A": "global-a", "B": "patch-b", "C": "patch-c"}
	for k, v := range want {
		if merged.Secrets[k] != v {
			t.Fatalf("merged[%q] = %q, want %q", k, merged.Secrets[k], v)
		}
	}
}

func TestSecretsLookup(t *testing.T) {
	sec := NewSecrets()
	sec.Secrets["KEY"] = "val"

	lk := SecretsLookup(sec)
	if v, ok := lk("KEY"); !ok || v != "val" {
		t.Fatalf("lookup = %q %v", v, ok)
	}
	if _, ok := lk("MISSING"); ok {
		t.Fatal("expected missing")
	}
}

func TestSetupCgroupFileLayout(t *testing.T) {
	root := t.TempDir()
	limits := Limits{MemoryMB: 128, CPUQuota: 25000, CPUPeriod: 100000, PIDsMax: 4}

	cgPath, err := SetupCgroup(root, "test-bridge", limits)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if !strings.HasSuffix(cgPath, filepath.Join("cognitiveos", "test-bridge")) {
		t.Fatalf("path = %q", cgPath)
	}

	if got := readFile(t, filepath.Join(cgPath, "memory.max")); got != "128M" {
		t.Fatalf("memory.max = %q", got)
	}
	if got := readFile(t, filepath.Join(cgPath, "cpu.max")); got != "25000 100000" {
		t.Fatalf("cpu.max = %q", got)
	}
	if got := readFile(t, filepath.Join(cgPath, "pids.max")); got != "4" {
		t.Fatalf("pids.max = %q", got)
	}
}

func TestJoinCgroup(t *testing.T) {
	root := t.TempDir()
	cgPath, err := SetupCgroup(root, "join-bridge", DefaultLimits())
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	pid := os.Getpid()
	if err := JoinCgroup(cgPath, pid); err != nil {
		t.Fatalf("join: %v", err)
	}
	got := readFile(t, filepath.Join(cgPath, "cgroup.procs"))
	if !strings.Contains(got, strconv.Itoa(pid)) {
		t.Fatalf("cgroup.procs = %q, want to contain pid %d", got, pid)
	}

	if err := JoinCgroup(cgPath, 0); err == nil {
		t.Fatal("expected error joining pid 0")
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return strings.TrimRight(string(data), "\n")
}
