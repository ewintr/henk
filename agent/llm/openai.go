package llm

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"go-mod.ewintr.nl/henk/agent/tool"
)

type OpenAI struct {
	client         *openai.Client
	provider       Provider
	modelName      string
	modelShortName string
	systemPrompt   string
}

func NewOpenAI(provider Provider, modelName, systemPrompt string) (*OpenAI, error) {
	m, ok := provider.Model(modelName)
	if !ok {
		return nil, fmt.Errorf("%w: could not find model %q in provider %q", ErrUnknownModel, modelName, provider.Name)
	}

	config := openai.DefaultConfig(provider.ApiKey)
	config.BaseURL = provider.BaseURL
	c := openai.NewClientWithConfig(config)
	return &OpenAI{
		client:         c,
		provider:       provider,
		modelName:      m.Name,
		modelShortName: m.ShortName,
		systemPrompt:   systemPrompt,
	}, nil
}

func (o *OpenAI) ModelInfo() (string, string, string) {
	return o.provider.Name, o.modelName, o.modelShortName
}

func (o *OpenAI) RunInference(ctx context.Context, tools []tool.Tool, conversation []Message) (Message, error) {
	openaiConv := make([]openai.ChatCompletionMessage, 0, len(conversation)+1)
	openaiConv = append(openaiConv, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: o.systemPrompt,
	})
	for _, msg := range conversation {
		for _, block := range msg.Content {
			var openaiMsg openai.ChatCompletionMessage
			switch block.Type {
			case ContentTypeText:
				role := ""
				switch msg.Role {
				case RoleAssistant:
					role = openai.ChatMessageRoleAssistant
				case RoleUser:
					role = openai.ChatMessageRoleUser
				default:
					return Message{}, fmt.Errorf("unknown message role: %s", msg.Role)
				}
				openaiMsg = openai.ChatCompletionMessage{
					Role:    role,
					Content: block.Text,
				}
			case ContentTypeToolUse:
				tu := block.ToolUse
				openaiMsg = openai.ChatCompletionMessage{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{
						{
							ID:   tu.ID,
							Type: openai.ToolTypeFunction,
							Function: openai.FunctionCall{
								Name:      tu.Name,
								Arguments: string(tu.Input),
							},
						},
					},
				}
			case ContentTypeToolResult:
				tr := block.ToolResult
				openaiMsg = openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    tr.Result,
					ToolCallID: tr.ID,
				}
			default:
				return Message{}, fmt.Errorf("unknown message content type: %s", block.Type)
			}
			openaiConv = append(openaiConv, openaiMsg)
		}
	}

	openaiTools := []openai.Tool{}
	for _, tool := range tools {
		openaiTools = append(openaiTools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.InputSchema(),
			},
		})
	}

	req := openai.ChatCompletionRequest{
		Model:    o.modelName,
		Messages: openaiConv,
		Tools:    openaiTools,
	}

	resp, err := o.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return Message{}, fmt.Errorf("ChatCompletion error: %v", err)
	}

	message := Message{
		Role:    RoleAssistant,
		Content: []ContentBlock{},
	}

	choice := resp.Choices[0]
	if choice.Message.Content != "" {
		message.Content = append(message.Content, ContentBlock{
			Type: ContentTypeText,
			Text: choice.Message.Content,
		})
	}

	for _, toolCall := range choice.Message.ToolCalls {
		if toolCall.Type == openai.ToolTypeFunction {
			message.Content = append(message.Content, ContentBlock{
				Type: ContentTypeToolUse,
				ToolUse: ToolUse{
					ID:    toolCall.ID,
					Name:  toolCall.Function.Name,
					Input: []byte(toolCall.Function.Arguments),
				},
			})
		}
	}

	return message, nil
}
