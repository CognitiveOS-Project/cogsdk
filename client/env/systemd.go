package env

import (
	"os"
	"os/exec"
	"strings"
)

// SystemdAvailable reports whether the system is running systemd as PID 1 by
// checking for the systemd runtime directory.
func SystemdAvailable() bool {
	_, err := os.Stat("/run/systemd/system")
	return err == nil
}

// UnitActive reports whether a systemd unit is in the "active" state using
// `systemctl is-active`. Returns false on any error (including non-systemd).
func UnitActive(unit string) bool {
	out, err := exec.Command("systemctl", "is-active", unit).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "active"
}
