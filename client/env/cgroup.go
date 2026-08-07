package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// CgroupRoot is the cgroup v2 mount point.
const CgroupRoot = "/sys/fs/cgroup"

// Limits describes cgroup v2 resource limits for an MCP server.
type Limits struct {
	MemoryMB    int64
	CPUQuota    int
	CPUPeriod   int
	PIDsMax     int
	IOReadMBps  int64
	IOWriteMBps int64
}

// DefaultLimits returns the standard bridge limits used by cognitiveosd.
func DefaultLimits() Limits {
	return Limits{
		MemoryMB:    512,
		CPUQuota:    25000,
		CPUPeriod:   100000,
		PIDsMax:     16,
		IOReadMBps:  10,
		IOWriteMBps: 5,
	}
}

// SetupCgroup creates a cognitiveos/<name> cgroup and applies the given limits.
// Returns the cgroup path.
func SetupCgroup(root, name string, limits Limits) (string, error) {
	cgPath := filepath.Join(root, "cognitiveos", name)
	if err := os.MkdirAll(cgPath, 0755); err != nil {
		return "", fmt.Errorf("mkdir cgroup: %w", err)
	}

	if err := writeCgroupValue(cgPath, "memory.max", fmt.Sprintf("%dM", limits.MemoryMB)); err != nil {
		return cgPath, fmt.Errorf("set memory.max: %w", err)
	}
	if err := writeCgroupValue(cgPath, "cpu.max", fmt.Sprintf("%d %d", limits.CPUQuota, limits.CPUPeriod)); err != nil {
		return cgPath, fmt.Errorf("set cpu.max: %w", err)
	}
	if err := writeCgroupValue(cgPath, "pids.max", strconv.Itoa(limits.PIDsMax)); err != nil {
		return cgPath, fmt.Errorf("set pids.max: %w", err)
	}
	ioRule := fmt.Sprintf("0:0 riops=max wiops=max rbps=%d wbps=%d\n", limits.IOReadMBps*1024*1024, limits.IOWriteMBps*1024*1024)
	if err := writeCgroupValue(cgPath, "io.max", ioRule); err != nil {
		return cgPath, fmt.Errorf("set io.max: %w", err)
	}

	return cgPath, nil
}

// JoinCgroup moves a pid into the cgroup at cgPath.
func JoinCgroup(cgPath string, pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid: %d", pid)
	}
	return writeCgroupValue(cgPath, "cgroup.procs", strconv.Itoa(pid))
}

func writeCgroupValue(cgPath, file, value string) error {
	return os.WriteFile(filepath.Join(cgPath, file), []byte(value), 0644)
}
