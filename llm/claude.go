package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"go-mod.ewintr.nl/henk/tool"
)

type Claude struct {
	client *anthropic.Client
}

func NewClaude() *Claude {
	c := anthropic.NewClient()
	return &Claude{
		client: &c,
	}
}

func (c *Claude) RunInference(ctx context.Context, tools []tool.Tool, conversation []Message) (Message, error) {
	antConv := make([]anthropic.MessageParam, 0, len(conversation))
	for _, msg := range conversation {
		for _, block := range msg.Content {
			var antBlock anthropic.ContentBlockParamUnion
			switch block.Type {
			case ContentTypeText:
				antBlock = anthropic.NewTextBlock(block.Text)
			case ContentTypeToolUse:
				tu := block.ToolUse
				antBlock = anthropic.NewToolUseBlock(tu.ID, tu.Input, tu.Name)
			case ContentTypeToolResult:
				tr := block.ToolResult
				antBlock = anthropic.NewToolResultBlock(tr.ID, tr.Result, tr.Error)
			default:
				return Message{}, fmt.Errorf("Error: unknown message content type: %s\n", block.Type)
			}
			switch msg.Role {
			case RoleAssistant:
				antConv = append(antConv, anthropic.NewAssistantMessage(antBlock))
			case RoleUser:
				antConv = append(antConv, anthropic.NewUserMessage(antBlock))
			default:
				return Message{}, fmt.Errorf("Error: unknown message role: %s\n", msg.Role)
			}
		}
	}

	antTools := []anthropic.ToolUnionParam{}
	for _, tool := range tools {
		antTools = append(antTools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name(),
				Description: anthropic.String(tool.Description()),
				InputSchema: tool.InputSchema(),
			},
		})
	}

	antMessage, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_7SonnetLatest,
		MaxTokens: int64(2048),
		Messages:  antConv,
		Tools:     antTools,
	})
	if err != nil {
		return Message{}, err
	}
	// return Message{}, fmt.Errorf("here...")

	message := Message{
		Role:    RoleAssistant,
		Content: []ContentBlock{},
	}
	for _, block := range antMessage.ToParam().Content {
		tp := block.GetType()
		switch *tp {
		case "text":
			text := block.GetText()
			message.Content = append(message.Content, ContentBlock{
				Type: ContentTypeText,
				Text: *text,
			})
		case "tool_use":
			id := block.GetID()
			name := block.GetName()
			inputAnyPtr := block.GetInput()
			inputAny := *inputAnyPtr
			input, ok := inputAny.(json.RawMessage)
			if !ok {
				return Message{}, fmt.Errorf("could not cast tool input to json.RawMessage")
			}
			message.Content = append(message.Content, ContentBlock{
				Type: ContentTypeToolUse,
				ToolUse: ToolUse{
					ID:    *id,
					Name:  *name,
					Input: input,
				},
			})
		default:
			return Message{}, fmt.Errorf("unknown content type: %s\n", *tp)
		}
	}

	return message, nil
}
