# Forge SDK

Go module (`github.com/mwantia/forge-sdk`) providing shared interfaces, gRPC implementations, and utilities for building Forge plugins.

## Overview

The SDK defines a plugin architecture where each plugin is an external process communicating over gRPC via [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin). Plugins implement a `Driver` interface that exposes one or more typed sub-plugins.

## Plugin Types

| Type | Constant | Description |
|------|----------|-------------|
| Provider | `PluginTypeProvider` | LLM provider (Ollama, Anthropic, etc.) |
| Memory | `PluginTypeMemory` | Memory/vector storage (sessions, embeddings) |
| Channel | `PluginTypeChannel` | Communication gateway (Discord, etc.) |
| Tools | `PluginTypeTools` | Tool calling bridge |
| Sandbox | `PluginTypeSandbox` | Isolated execution environment (Docker, SSH, etc.) |

## Package Structure

```
pkg/
├── plugins/         # Core interfaces
│   └── grpc/        # gRPC transport layer
│       ├── driver/
│       ├── provider/
│       ├── memory/
│       ├── channel/
│       ├── tools/
│       └── sandbox/
├── errors/          # Shared error values
├── log/             # hclog integration and colored output wrapper
└── metrics/         # Prometheus metrics definitions
```

## Implementing a Plugin

Every plugin implements `plugins.Driver`. Sub-plugins are optional — return `errors.ErrPluginNotSupported` for capabilities your driver does not provide.

```go
package main

import (
    "github.com/hashicorp/go-hclog"
    "github.com/mwantia/forge-sdk/pkg/plugins"
    "github.com/mwantia/forge-sdk/pkg/plugins/grpc"
)

type MyDriver struct{}

func (d *MyDriver) GetPluginInfo() plugins.PluginInfo {
    return plugins.PluginInfo{Name: "my-plugin", Version: "0.1.0"}
}

// ... implement remaining Driver methods

func main() {
    grpc.Serve(func(log hclog.Logger) plugins.Driver {
        return &MyDriver{}
    })
}
```

Use `UnimplementedProviderPlugin`, `UnimplementedMemoryPlugin`, etc. as embedded stubs for sub-plugins you don't support.

### Registering a Plugin

Call `plugins.Register` in an `init()` function to make a driver discoverable:

```go
func init() {
    plugins.Register("my-plugin", "Short description", func(log hclog.Logger) plugins.Driver {
        return &MyDriver{}
    })
}
```

## Key Interfaces

### `Driver`

```go
type Driver interface {
    Lifecycle
    OpenDriver(ctx context.Context) error
    CloseDriver(ctx context.Context) error
    ConfigDriver(ctx context.Context, config PluginConfig) error
    GetProviderPlugin(ctx context.Context) (ProviderPlugin, error)
    GetMemoryPlugin(ctx context.Context) (MemoryPlugin, error)
    GetChannelPlugin(ctx context.Context) (ChannelPlugin, error)
    GetToolsPlugin(ctx context.Context) (ToolsPlugin, error)
    GetSandboxPlugin(ctx context.Context) (SandboxPlugin, error)
}
```

### `ProviderPlugin`

Streaming chat, embeddings, and model management:

```go
Chat(ctx, messages []ChatMessage, tools []ToolCall, model string) (ChatStream, error)
Embed(ctx, content []string, model string) ([][]float32, error)
ListModels(ctx) ([]*Model, error)
```

### `MemoryPlugin`

Session-scoped vector storage:

```go
StoreResource(ctx, sessionID, content string, metadata map[string]string) (*MemoryResource, error)
RetrieveResource(ctx, sessionID, query string, limit int, filter map[string]string) ([]*MemoryResource, error)
CreateSession(ctx) (*MemorySession, error)
```

### `ToolsPlugin`

Tool discovery and execution with streaming output:

```go
ListTools(ctx, filter ListToolsFilter) (*ListToolsResponse, error)
Execute(ctx, req ExecuteRequest) (*ExecuteResponse, error)
ExecuteStream(ctx, req ExecuteRequest) (<-chan ExecuteChunk, error)
```

### `SandboxPlugin`

Isolated execution with filesystem access:

```go
CreateSandbox(ctx, spec SandboxSpec) (*SandboxHandle, error)
Execute(ctx, req SandboxExecRequest) (<-chan SandboxExecChunk, error)
CopyIn(ctx, id, hostSrc, sandboxDst string) error
CopyOut(ctx, id, sandboxSrc, hostDst string) error
```

## gRPC Transport

The `pkg/plugins/grpc` package wires each plugin interface to generated protobuf/gRPC code. Plugin handshake uses:

- Protocol version: `2`
- Magic cookie key: `FORGE_PLUGIN`
- Magic cookie value: `forge`

Server entry points in `grpc/serve.go`:

| Function | Description |
|----------|-------------|
| `Serve(factory)` | Start server with default logger |
| `ServeWithLogger(factory, logger)` | Start with custom hclog logger |
| `ServeContext(factory)` | Start with context propagation |
| `ServeContextWithLogger(factory, logger)` | Start with both |

## Utilities

### Errors (`pkg/errors`)

```go
errors.ErrPluginNotYetImplemented
errors.ErrPluginNotSupported
errors.ErrPluginCapabilityNotSupported
errors.ErrSkillNotFound
errors.ErrInvalidSkillPath
```

The `Errors` type is a thread-safe error collector:

```go
var errs errors.Errors
errs.Add(err1)
errs.Add(err2)
return errs.Errors() // joined error or nil
```

### Logging (`pkg/log`)

`LogWrapper` adds timestamps and ANSI color codes to any `io.Writer`:

```go
w := log.LogWrapper(os.Stderr, true)
```

`HcLogTagProcessor` integrates with `github.com/mwantia/fabric` for automatic logger injection using `fabric:"logger"` struct tags.

### Metrics (`pkg/metrics`)

Prometheus metrics registered globally:

```go
metrics.ServerHttpRequestsTotal          // CounterVec — method, address, path, status
metrics.ServerHttpRequestsDurationSeconds // HistogramVec — method, address, path
```

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/hashicorp/go-plugin` | Plugin process lifecycle and transport |
| `github.com/hashicorp/go-hclog` | Structured logging |
| `google.golang.org/grpc` | gRPC framework |
| `google.golang.org/protobuf` | Protocol buffers |
| `github.com/mwantia/fabric` | Dependency injection / tag processors |
| `github.com/prometheus/client_golang` | Prometheus metrics |
