package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go-mod.ewintr.nl/henk/tool"
)

type Ollama struct {
	baseURL      string
	model        string
	systemPrompt string
	contextSize  int
	client       *http.Client
}

func NewOllama(baseURL, model, systemPrompt string, contextSize int) *Ollama {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &Ollama{
		baseURL:      baseURL,
		model:        model,
		systemPrompt: systemPrompt,
		contextSize:  contextSize,
		client:       &http.Client{},
	}
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaTool struct {
	Type     string             `json:"type"`
	Function ollamaToolFunction `json:"function"`
}

type ollamaToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ollamaToolCall struct {
	Function ollamaToolCallFunction `json:"function"`
}

type ollamaToolCallFunction struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ollamaChatRequestOptions struct {
	NumCtx int `json:"num_ctx,omitempty"`
}

type ollamaChatRequest struct {
	Model    string                   `json:"model"`
	Messages []ollamaMessage          `json:"messages"`
	Tools    []ollamaTool             `json:"tools,omitempty"`
	Stream   bool                     `json:"stream"`
	Options  ollamaChatRequestOptions `json:"options,omitempty"`
}

type ollamaChatResponse struct {
	Message    ollamaResponseMessage `json:"message"`
	Done       bool                  `json:"done"`
	DoneReason string                `json:"done_reason,omitempty"`
}

type ollamaResponseMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
}

func (o *Ollama) RunInference(ctx context.Context, tools []tool.Tool, conversation []Message) (Message, error) {
	// Convert internal messages to Ollama format
	ollamaMessages := make([]ollamaMessage, 0, len(conversation)+1)
	ollamaMessages = append(ollamaMessages, ollamaMessage{
		Role:    "system",
		Content: o.systemPrompt,
	})

	for _, msg := range conversation {
		if msg.Role == RoleUser || msg.Role == RoleAssistant {
			// Convert content blocks to appropriate format
			var content strings.Builder
			var toolResults []string

			for _, block := range msg.Content {
				switch block.Type {
				case ContentTypeText:
					content.WriteString(block.Text)
				case ContentTypeToolResult:
					toolResults = append(toolResults, fmt.Sprintf("Tool %s result: %s", block.ToolResult.ID, block.ToolResult.Result))
				}
			}

			// Add tool results to content if any
			if len(toolResults) > 0 {
				if content.Len() > 0 {
					content.WriteString("\n")
				}
				content.WriteString(strings.Join(toolResults, "\n"))
			}

			if content.Len() > 0 {
				ollamaMessages = append(ollamaMessages, ollamaMessage{
					Role:    string(msg.Role),
					Content: content.String(),
				})
			}
		}
	}

	// Convert tools to Ollama format
	ollamaTools := make([]ollamaTool, 0, len(tools))
	for _, t := range tools {
		schema := t.InputSchema()
		// Convert jsonschema.Schema to map[string]interface{}
		schemaBytes, err := json.Marshal(schema)
		if err != nil {
			return Message{}, fmt.Errorf("failed to marshal tool schema: %w", err)
		}
		var schemaMap map[string]interface{}
		if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
			return Message{}, fmt.Errorf("failed to unmarshal tool schema: %w", err)
		}

		ollamaTools = append(ollamaTools, ollamaTool{
			Type: "function",
			Function: ollamaToolFunction{
				Name:        t.Name(),
				Description: t.Description(),
				Parameters:  schemaMap,
			},
		})
	}

	// Create request
	request := ollamaChatRequest{
		Model:    o.model,
		Messages: ollamaMessages,
		Tools:    ollamaTools,
		Options: ollamaChatRequestOptions{
			NumCtx: o.contextSize,
		},
	}

	// Make HTTP request
	resp, err := o.makeRequest(ctx, request)
	if err != nil {
		return Message{}, err
	}

	// Convert response to internal format
	return o.convertResponse(resp, tools)
}

func (o *Ollama) makeRequest(ctx context.Context, request ollamaChatRequest) (*ollamaChatResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func (o *Ollama) convertResponse(resp *ollamaChatResponse, tools []tool.Tool) (Message, error) {
	message := Message{
		Role:    RoleAssistant,
		Content: make([]ContentBlock, 0),
	}

	// Add text content if present
	if resp.Message.Content != "" {
		message.Content = append(message.Content, ContentBlock{
			Type: ContentTypeText,
			Text: resp.Message.Content,
		})
	}

	// Handle tool calls
	for i, toolCall := range resp.Message.ToolCalls {
		toolUse := ToolUse{
			ID:    fmt.Sprintf("tool_call_%d", i),
			Name:  toolCall.Function.Name,
			Input: toolCall.Function.Arguments,
		}

		message.Content = append(message.Content, ContentBlock{
			Type:    ContentTypeToolUse,
			ToolUse: toolUse,
		})
	}

	return message, nil
}
