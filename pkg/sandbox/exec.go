package sandbox

import "io"

// ExecRequest describes a script execution within the sandbox.
type ExecRequest struct {
	Path        string // relative path to the file within sandbox root
	Args        []string
	Environment []string // additional KEY=VALUE env vars
	Stdin       io.Reader

	// 0 = use Config.MaxOutputBytes
	MaxOutputBytes int
}

// ExecResult holds the outcome of a script execution.
type ExecResult struct {
	ExitCode  int
	Stdout    string
	Stderr    string
	Truncated bool
}
