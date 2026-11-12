package target

import (
	"fmt"
	"sync"
	"time"
)

// Target represents a deployment target environment.
type Target struct {
	// Name is the unique identifier for the target.
	Name string
	// Endpoint is the network endpoint for the target.
	Endpoint string
	// Region is the geographic region of the target.
	Region string
	// Status is the current provisioning status.
	Status string
}

// Provisioner provisions target environments.
type Provisioner struct {
	mu      sync.RWMutex
	targets map[string]*Target
}

// NewProvisioner creates a new Target Provisioner.
func NewProvisioner() *Provisioner {
	return &Provisioner{
		targets: make(map[string]*Target),
	}
}

// RegisterTarget registers a new target with the provisioner.
func (p *Provisioner) RegisterTarget(name, endpoint, region string) {
	p.targets[name] = &Target{
		Name:     name,
		Endpoint: endpoint,
		Region:   region,
		Status:   "pending",
	}
}

// Provision provisions a target by name.
func (p *Provisioner) Provision(name string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	t, ok := p.targets[name]
	if !ok {
		return fmt.Errorf("target %q not found", name)
	}

	t.Status = "provisioning"
	p.targets[name] = t

	// Simulate provisioning delay
	time.Sleep(50 * time.Millisecond)

	t.Status = "active"
	p.targets[name] = t
	return nil
}

// GetTarget returns a target by name.
func (p *Provisioner) GetTarget(name string) (*Target, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	t, ok := p.targets[name]
	if !ok {
		return nil, fmt.Errorf("target %q not found", name)
	}
	return t, nil
}

// ListTargets returns all registered targets.
func (p *Provisioner) ListTargets() []*Target {
	p.mu.RLock()
	defer p.mu.RUnlock()

	targets := make([]*Target, 0, len(p.targets))
	for _, t := range p.targets {
		targets = append(targets, t)
	}
	return targets
}