package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go-mod.ewintr.nl/henk/agent/tool"
)

var (
	ErrUnknownModel = errors.New("unknown model")
)

type LLM interface {
	ModelInfo() (provider, model, short string)
	RunInference(ctx context.Context, tools []tool.Tool, conversation []Message) (Message, error)
}

type ContentType string

const (
	ContentTypeText       ContentType = "text"
	ContentTypeToolUse    ContentType = "tool_use"
	ContentTypeToolResult ContentType = "tool_result"
)

type Role string

const (
	RoleAssistant Role = "assistant"
	RoleUser      Role = "user"
)

type ToolUse struct {
	ID    string
	Name  string
	Input json.RawMessage
}

type ToolResult struct {
	ID     string
	Result string
	Error  bool
}

type ContentBlock struct {
	ID         string
	Type       ContentType
	Text       string
	ToolUse    ToolUse
	ToolResult ToolResult
}

type Message struct {
	Content []ContentBlock
	Role    Role
}

type Conversation struct{}

type Model struct {
	Name         string `toml:"name"`
	ShortName    string `toml:"short_name"`
	Default      bool   `toml:"default"`
	ContextSize  int    `toml:"context_size"`
	ThinkingMode bool   `toml:"thinking_mode"`
}

type Provider struct {
	Type      string `toml:"type"`
	BaseURL   string `toml:"base_url"`
	ApiKey    string
	ApiKeyEnv string  `toml:"api_key_env"`
	Name      string  `toml:"name"`
	Models    []Model `toml:"models"`
}

func (p Provider) Model(name string) (Model, bool) {
	for _, m := range p.Models {
		if m.Name == name || m.ShortName == name {
			return m, true
		}
	}

	return Model{}, false
}

func NewLLM(provider Provider, modelName, systemPrompt string) (LLM, error) {
	switch provider.Type {
	case "claude":
		return NewClaude(provider, modelName, systemPrompt)
	case "openai":
		return NewOpenAI(provider, modelName, systemPrompt)
	case "ollama":
		return NewOllama(provider, modelName, systemPrompt)
	default:
		return nil, fmt.Errorf("unknown provider type: %s", provider.Type)
	}
}
