package plugins

import (
	"context"
	"io"

	"github.com/mwantia/forge-sdk/pkg/errors"
)

// ChatStream is returned by Chat for consuming a streaming response.
// Recv returns the next chunk; io.EOF signals the stream is complete.
// Close must always be called to release resources.
type ChatStream interface {
	Recv() (*ChatChunk, error)
	Close() error
}

// ProviderPlugin acts as LLM provider to "provide" access to endpoints like Ollama, Anthropic, etc.
type ProviderPlugin interface {
	BasePlugin

	// Chat sends messages and returns a stream of response chunks.
	Chat(ctx context.Context, messages []ChatMessage, tools []ToolCall, model *Model) (ChatStream, error)

	Embed(ctx context.Context, content string, model *Model) ([][]float32, error)

	ListModels(ctx context.Context) ([]*Model, error)
	CreateModel(ctx context.Context, modelName string, template *ModelTemplate) (*Model, error)
	GetModel(ctx context.Context, name string) (*Model, error)
	DeleteModel(ctx context.Context, name string) (bool, error)
}

type Model struct {
	ModelName          string         `json:"model_name,omitempty"`
	Dimension          int            `json:"dimension,omitempty"`
	Temperature        float64        `json:"temperature,omitempty"`
	Template           *ModelTemplate `json:"template,omitempty"`
	Metadata           map[string]any `json:"metadata,omitempty"`
	CostPerInputToken  float64        `json:"cost_per_input_token,omitempty"`
	CostPerOutputToken float64        `json:"cost_per_output_token,omitempty"`
}

// TokenUsage tracks token consumption and optional cost for a single inference call.
type TokenUsage struct {
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	TotalTokens  int     `json:"total_tokens"`
	InputCost    float64 `json:"input_cost,omitempty"`
	OutputCost   float64 `json:"output_cost,omitempty"`
	TotalCost    float64 `json:"total_cost,omitempty"`
}

// Add combines another TokenUsage into this one (used to accumulate session totals).
func (u *TokenUsage) Add(other *TokenUsage) {
	if other == nil {
		return
	}
	u.InputTokens += other.InputTokens
	u.OutputTokens += other.OutputTokens
	u.TotalTokens += other.TotalTokens
	u.InputCost += other.InputCost
	u.OutputCost += other.OutputCost
	u.TotalCost += other.TotalCost
}

type ModelTemplate struct {
	BaseModel      string         `json:"base_model"`
	PromptTemplate string         `json:"prompt_template,omitempty"`
	System         string         `json:"system,omitempty"`
	Parameters     map[string]any `json:"parameters,omitempty"`
}

// ChatChunk is a single streamed piece of a response.
// Delta carries incremental text. ToolCalls, Done, and Usage are only set on the final chunk.
type ChatChunk struct {
	ID        string         `json:"id,omitempty"`
	Role      string         `json:"role,omitempty"`
	Delta     string         `json:"delta"`
	Done      bool           `json:"done"`
	ToolCalls []ChatToolCall `json:"tool_calls,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Usage     *TokenUsage    `json:"usage,omitempty"`
}

// ChatResult is the fully-assembled response after draining a ChatStream.
type ChatResult struct {
	ID        string         `json:"id"`
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	ToolCalls []ChatToolCall `json:"tool_calls,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Usage     *TokenUsage    `json:"usage,omitempty"`
}

// CollectStream drains a ChatStream into a single ChatResult.
func CollectStream(stream ChatStream) (*ChatResult, error) {
	defer stream.Close()
	result := &ChatResult{}
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			return result, nil
		}
		if err != nil {
			return nil, err
		}
		if result.ID == "" {
			result.ID = chunk.ID
		}
		if result.Role == "" {
			result.Role = chunk.Role
		}
		result.Content += chunk.Delta
		result.ToolCalls = append(result.ToolCalls, chunk.ToolCalls...)
		if chunk.Done {
			result.Metadata = chunk.Metadata
			result.Usage = chunk.Usage
			return result, nil
		}
	}
}

type ChatToolCall struct {
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type ChatMessage struct {
	Role      string                `json:"role"`
	Content   string                `json:"content"`
	ToolCalls *ChatMessageToolCalls `json:"tool_calls,omitempty"`
}

type ChatMessageToolCalls struct {
	ID        string         `json:"id,omitempty"`
	Name      string         `json:"name,omitempty"`
	ToolCalls []ChatToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  ToolParameters `json:"parameters"`
}

// UnimplementedProviderPlugin can be embedded to satisfy ProviderPlugin with default
// implementations that return errors.ErrPluginCapabilityNotSupported.
type UnimplementedProviderPlugin struct{}

func (UnimplementedProviderPlugin) GetLifecycle() Lifecycle { return nil }

func (UnimplementedProviderPlugin) Chat(_ context.Context, _ []ChatMessage, _ []ToolCall, _ *Model) (ChatStream, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedProviderPlugin) Embed(_ context.Context, _ string, _ *Model) ([][]float32, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedProviderPlugin) ListModels(_ context.Context) ([]*Model, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedProviderPlugin) CreateModel(_ context.Context, _ string, _ *ModelTemplate) (*Model, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedProviderPlugin) GetModel(_ context.Context, _ string) (*Model, error) {
	return nil, errors.ErrPluginCapabilityNotSupported
}

func (UnimplementedProviderPlugin) DeleteModel(_ context.Context, _ string) (bool, error) {
	return false, errors.ErrPluginCapabilityNotSupported
}

var _ ProviderPlugin = (*UnimplementedProviderPlugin)(nil)
