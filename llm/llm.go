package llm

import (
	"context"
	"encoding/json"

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
