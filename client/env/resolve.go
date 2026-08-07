/*
Package env provides environment inspection and resolution: cgroup placement,
systemd unit state, and config/env resolution (including the ${VAR} secret
resolution pattern from ADR-010).
*/
package env

import (
	"errors"
	"regexp"
)

// ErrUndefined is the sentinel wrapped by UndefinedError.
var ErrUndefined = errors.New("env: undefined variable")

var placeholderPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// Lookup resolves a variable name to a value and whether it is defined.
type Lookup func(name string) (string, bool)

// EnvLookup resolves from the process environment.
var EnvLookup Lookup = func(name string) (string, bool) {
	return osLookupEnv(name)
}

// Resolve replaces ${VAR} placeholders in data using lookup. It returns an
// error naming the first undefined variable if any placeholder cannot be
// resolved.
func Resolve(data []byte, lookup Lookup) ([]byte, error) {
	if lookup == nil {
		lookup = EnvLookup
	}

	seen := map[string]bool{}
	for _, m := range placeholderPattern.FindAllSubmatch(data, -1) {
		name := string(m[1])
		if seen[name] {
			continue
		}
		seen[name] = true
		if _, ok := lookup(name); !ok {
			return nil, &UndefinedError{Name: name}
		}
	}

	out := placeholderPattern.ReplaceAllStringFunc(string(data), func(match string) string {
		name := placeholderPattern.FindStringSubmatch(match)
		if len(name) < 2 {
			return match
		}
		v, _ := lookup(name[1])
		return v
	})
	return []byte(out), nil
}

// UndefinedError reports a placeholder that referenced an undefined variable.
type UndefinedError struct {
	Name string
}

func (e *UndefinedError) Error() string {
	return "undefined variable: ${" + e.Name + "}"
}

func (e *UndefinedError) Unwrap() error {
	return ErrUndefined
}
