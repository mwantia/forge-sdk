package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// TinySandbox is an os.Root-backed Sandbox that confines file access to a single directory tree.
// Script execution uses the root as the working directory; os.Root prevents symlink escapes
// and path traversal during path validation before exec.
type TinySandbox struct {
	root     *os.Root
	config   Config
	profiles map[string]ExecProfile // keyed by file extension (without leading dot)
}

// NewTinySandbox opens dir as the sandbox root with the provided Config.
// The preset ExecProfiles (bash, python3, node) are registered by default.
func NewTinySandbox(dir string, cfg Config) (*TinySandbox, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("open sandbox root %q: %w", dir, err)
	}

	t := &TinySandbox{
		root:     root,
		config:   cfg,
		profiles: make(map[string]ExecProfile),
	}

	for _, p := range ExecProfiles {
		for _, ext := range p.Extensions {
			t.profiles[ext] = p
		}
	}

	return t, nil
}

func (t *TinySandbox) hasCap(c Capability) bool {
	return slices.Contains(t.config.Caps, c)
}

// Root returns the absolute path of the sandbox root directory.
func (t *TinySandbox) Root() string {
	return t.root.Name()
}

// Close releases the underlying file descriptor.
func (t *TinySandbox) Close() error {
	return t.root.Close()
}

// RegisterProfile adds p to the profile registry, keyed by each of its extensions.
// An existing entry for the same extension is overwritten.
func (t *TinySandbox) RegisterProfile(p ExecProfile) error {
	if p.Name == "" {
		return fmt.Errorf("profile name must not be empty")
	}
	if len(p.Extensions) == 0 {
		return fmt.Errorf("profile %q has no extensions", p.Name)
	}
	for _, ext := range p.Extensions {
		t.profiles[ext] = p
	}
	return nil
}

// UnregisterProfile removes all extension entries that map to the named profile.
// Returns true if at least one entry was removed.
func (t *TinySandbox) UnregisterProfile(name string) (bool, error) {
	removed := false
	for ext, p := range t.profiles {
		if p.Name == name {
			delete(t.profiles, ext)
			removed = true
		}
	}
	return removed, nil
}

// ReadFile reads the named file within the sandbox.
func (t *TinySandbox) ReadFile(_ context.Context, path string) ([]byte, error) {
	if !t.hasCap(CapRead) {
		return nil, fmt.Errorf("capability %q not granted", CapRead)
	}

	return t.root.ReadFile(path)
}

// WriteFile writes data to path, creating or truncating the file.
func (t *TinySandbox) WriteFile(_ context.Context, path string, data []byte) error {
	if !t.hasCap(CapWrite) {
		return fmt.Errorf("capability %q not granted", CapWrite)
	}

	f, err := t.root.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(data)

	return err
}

// MkdirAll creates path and any missing parents within the sandbox.
func (t *TinySandbox) MkdirAll(_ context.Context, path string) error {
	if !t.hasCap(CapCreate) {
		return fmt.Errorf("capability %q not granted", CapCreate)
	}

	return t.root.MkdirAll(path, 0755)
}

// Remove removes path and all its children from the sandbox.
func (t *TinySandbox) Remove(_ context.Context, path string) error {
	if !t.hasCap(CapDelete) {
		return fmt.Errorf("capability %q not granted", CapDelete)
	}

	return t.root.RemoveAll(path)
}

// ReadDir lists the immediate children of path within the sandbox.
func (t *TinySandbox) ReadDir(_ context.Context, path string) ([]FileInfo, error) {
	if !t.hasCap(CapList) {
		return nil, fmt.Errorf("capability %q not granted", CapList)
	}

	entries, err := fs.ReadDir(t.root.FS(), path)
	if err != nil {
		return nil, err
	}

	result := make([]FileInfo, 0, len(entries))
	for _, e := range entries {
		fi := FileInfo{Name: e.Name(), IsDir: e.IsDir()}

		if info, err := e.Info(); err == nil {
			fi.Size = info.Size()
			fi.ModTime = info.ModTime()
		}
		result = append(result, fi)
	}

	return result, nil
}

// ExecProfile runs req.Path (relative to the sandbox root) using the profile
// registered for its file extension. os.Root.Lstat validates the path before
// execution — symlink escapes and path traversal are rejected at the OS level.
func (t *TinySandbox) ExecProfile(ctx context.Context, req ExecRequest) (ExecResult, error) {
	if !t.hasCap(CapExec) {
		return ExecResult{}, fmt.Errorf("capability %q not granted", CapExec)
	}

	// os.Root.Lstat rejects symlink escapes and traversal before we ever touch exec.
	if _, err := t.root.Lstat(req.Path); err != nil {
		return ExecResult{}, fmt.Errorf("script %q: %w", req.Path, err)
	}

	ext := strings.TrimPrefix(filepath.Ext(req.Path), ".")
	profile, ok := t.profiles[ext]
	if !ok {
		return ExecResult{}, fmt.Errorf("no exec profile registered for extension %q", ext)
	}

	absScript := filepath.Join(t.root.Name(), filepath.Clean(req.Path))

	timeout := t.config.RuntimeTimeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	env := make([]string, 0, len(t.config.InheritEnvironment)+len(req.Environment))
	for _, key := range t.config.InheritEnvironment {
		if val, ok := os.LookupEnv(key); ok {
			env = append(env, key+"="+val)
		}
	}
	env = append(env, req.Environment...)

	cmd := profile.Cmd(execCtx, append([]string{absScript}, req.Args...), env)
	cmd.Dir = t.root.Name()
	cmd.Stdin = req.Stdin

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	maxBytes := req.MaxOutputBytes
	if maxBytes <= 0 {
		maxBytes = t.config.MaxOutputBytes
	}
	if maxBytes <= 0 {
		maxBytes = 32768
	}

	outStr := stdout.String()
	errStr := stderr.String()

	return ExecResult{
		ExitCode:  exitCode,
		Stdout:    truncate(outStr, maxBytes),
		Stderr:    truncate(errStr, maxBytes),
		Truncated: len(outStr) > maxBytes || len(errStr) > maxBytes,
	}, nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}

	return s[:max]
}

var _ Sandbox = (*TinySandbox)(nil)
