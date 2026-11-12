package builder

import (
	"fmt"
	"os"
	"os/exec"
)

// Builder builds CognitiveOS container images and toolchains.
//
// Builder wraps common build operations such as Docker builds, OCI image
// creation, and toolchain compilation into a reusable Go package.
type Builder struct {
	// WorkDir is the working directory for build operations.
	WorkDir string
}

// NewBuilder creates a new Builder with the given working directory.
// If workDir is empty, the current working directory is used.
func NewBuilder(workDir string) *Builder {
	if workDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			wd = "."
		}
		workDir = wd
	}
	return &Builder{WorkDir: workDir}
}

// BuildImage builds a container image from a Dockerfile.
func (b *Builder) BuildImage(dockerfile, tag string) error {
	cmd := exec.Command("docker", "build", "-t", tag, "-f", dockerfile, b.WorkDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}
	return nil
}