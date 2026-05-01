package plugins

const (
	// A provider plugin acts as LLM provider to "provide" access to endpoints like Ollama, Anthropic, etc.
	PluginTypeProvider string = "provider"
	// A resource plugin owns namespaced resource storage and retrieval
	// (formerly memory) — used by adapters like OpenViking.
	PluginTypeResource string = "resource"
	// A channel plugin acts as communication gateway for endpoints like Discord.
	PluginTypeChannel string = "channel"
	// A tools plugin acts as bridge (or summary of embedded tools) for tool calling.
	PluginTypeTools string = "tools"
	// A sandbox plugin provides an isolated execution environment for running scripts and tools.
	PluginTypeSandbox string = "sandbox"
)
