package plugins

import (
	"context"

	"github.com/mwantia/forge-sdk/pkg/errors"
)

// DriverCapabilities describes what a driver supports.
type DriverCapabilities struct {
	Types    []string
	Provider *ProviderCapabilities
	Memory   *MemoryCapabilities
	Channel  *ChannelCapabilities
	Tools    *ToolsCapabilities
	Sandbox  *SandboxCapabilities
}

type ProviderCapabilities struct {
	SupportsStreaming bool
	SupportsVision    bool
}

type MemoryCapabilities struct {
	SupportsVectorSearch bool
	SupportSessions      bool
	MaxContextSize       int
}

type ChannelCapabilities struct {
	SupportsDirectMessages bool
	SupportsThreads        bool
}

type ToolsCapabilities struct {
	SupportsAsyncExecution bool
}

// PluginInfo describes plugin metadata at the driver level.
type PluginInfo struct {
	Name        string `json:"name"`
	Author      string `json:"author"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// PluginConfig holds driver configuration as a generic map.
type PluginConfig struct {
	ConfigMap map[string]any `json:"-"`
}

// Lifecycle provides access to driver-level lifecycle checks.
// Plugins use this to reference back to their parent driver.
type Lifecycle interface {
	GetPluginInfo() PluginInfo
	ProbePlugin(ctx context.Context) (bool, error)
	GetCapabilities(ctx context.Context) (*DriverCapabilities, error)
}

// Driver is the main interface that plugins implement.
// A single driver can support multiple plugin types simultaneously.
type Driver interface {
	Lifecycle

	// Lifecycle management
	OpenDriver(ctx context.Context) error
	CloseDriver(ctx context.Context) error

	// Configuration
	ConfigDriver(ctx context.Context, config PluginConfig) error

	// Plugin type accessors - return implementations only if supported
	GetProviderPlugin(ctx context.Context) (ProviderPlugin, error)
	GetMemoryPlugin(ctx context.Context) (MemoryPlugin, error)
	GetChannelPlugin(ctx context.Context) (ChannelPlugin, error)
	GetToolsPlugin(ctx context.Context) (ToolsPlugin, error)
	GetSandboxPlugin(ctx context.Context) (SandboxPlugin, error)
}

type UnimplementedDriver struct{}

// CloseDriver implements [Driver].
func (u *UnimplementedDriver) CloseDriver(ctx context.Context) error {
	return errors.ErrPluginCapabilityNotSupported
}

// ConfigDriver implements [Driver].
func (u *UnimplementedDriver) ConfigDriver(ctx context.Context, config PluginConfig) error {
	return errors.ErrPluginCapabilityNotSupported
}

// GetCapabilities implements [Driver].
func (u *UnimplementedDriver) GetCapabilities(ctx context.Context) (*DriverCapabilities, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

// GetChannelPlugin implements [Driver].
func (u *UnimplementedDriver) GetChannelPlugin(ctx context.Context) (ChannelPlugin, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

// GetMemoryPlugin implements [Driver].
func (u *UnimplementedDriver) GetMemoryPlugin(ctx context.Context) (MemoryPlugin, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

// GetPluginInfo implements [Driver].
func (u *UnimplementedDriver) GetPluginInfo() PluginInfo {
	return PluginInfo{}
}

// GetProviderPlugin implements [Driver].
func (u *UnimplementedDriver) GetProviderPlugin(ctx context.Context) (ProviderPlugin, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

// GetSandboxPlugin implements [Driver].
func (u *UnimplementedDriver) GetSandboxPlugin(ctx context.Context) (SandboxPlugin, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

// GetToolsPlugin implements [Driver].
func (u *UnimplementedDriver) GetToolsPlugin(ctx context.Context) (ToolsPlugin, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

// OpenDriver implements [Driver].
func (u *UnimplementedDriver) OpenDriver(ctx context.Context) error {
	return errors.ErrPluginCapabilityNotSupported
}

// ProbePlugin implements [Driver].
func (u *UnimplementedDriver) ProbePlugin(ctx context.Context) (bool, error) {
	return false, errors.ErrPluginCapabilityNotSupported
}

var _ Driver = (*UnimplementedDriver)(nil)
