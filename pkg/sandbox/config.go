package sandbox

import "time"

// Capability gates a class of operations within the sandbox.
type Capability string

const (
	CapRead   Capability = "read"
	CapWrite  Capability = "write"
	CapCreate Capability = "create"
	CapDelete Capability = "delete"
	CapList   Capability = "list"
	CapExec   Capability = "exec"
)

// Config controls what a Sandbox instance may do.
type Config struct {
	Caps               []Capability
	RuntimeTimeout     time.Duration
	MaxOutputBytes     int
	InheritEnvironment []string // env var names inherited from the parent process
}

// DefaultConfig returns a permissive config suitable for trusted workloads.
func DefaultConfig() Config {
	return Config{
		Caps:               []Capability{CapRead, CapWrite, CapCreate, CapDelete, CapList, CapExec},
		RuntimeTimeout:     60 * time.Second,
		MaxOutputBytes:     32768,
		InheritEnvironment: []string{"HOME", "PATH", "TZ", "LANG"},
	}
}
