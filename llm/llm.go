package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"go-mod.ewintr.nl/henk/tool"
)

type LLM interface {
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
	Name        string `toml:"name"`
	Default     bool   `toml:"default"`
	ContextSize int    `toml:"context_size"`
}

type Provider struct {
	Type      string  `toml:"type"`
	BaseURL   string  `toml:"base_url"`
	ApiKeyEnv string  `toml:"api_key_env"`
	Models    []Model `toml:"models"`
}

func (p Provider) DefaultModel() Model {
	var m Model
	// Find default model
	for _, model := range p.Models {
		if model.Default {
			m = model
		}
	}

	// If no default, return first model
	if m.Name == "" && len(p.Models) > 0 {
		m = p.Models[0]
	}

	if m.ContextSize == 0 {
		m.ContextSize = 8096
	}

	return m
}

func NewLLM(provider Provider, systemPrompt string) (LLM, error) {
	model := provider.DefaultModel()
	switch provider.Type {
	case "claude":
		return NewClaude(model.Name, systemPrompt), nil
	case "openai":
		var apiKey string
		if provider.ApiKeyEnv != "" {
			val, ok := os.LookupEnv(provider.ApiKeyEnv)
			if ok {
				apiKey = val
			}
		}
		return NewOpenAI(provider.BaseURL, apiKey, model.Name, systemPrompt), nil
	case "ollama":
		return NewOllama(provider.BaseURL, model.Name, systemPrompt, model.ContextSize), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", provider.Type)
	}
}
