package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"go-mod.ewintr.nl/henk/tool"
)

type Agent struct {
	client *anthropic.Client
	tools  []tool.Tool
	done   chan bool
}

func New(client *anthropic.Client, tools []tool.Tool) *Agent {
	return &Agent{
		client: client,
		tools:  tools,
		done:   make(chan bool),
	}
}

func (a *Agent) Run() error {
	go a.converse()

	for {
		select {
		case <-a.done:
			fmt.Println("Bye!")
			return nil
		}
	}

}

func (a *Agent) converse() error {
	ctx := context.Background()
	conversation := []anthropic.MessageParam{}
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Chat with Henk (use '/quit' to quit)")

	readUserInput := true
	for {
		if readUserInput {
			fmt.Print("\u001b[94mYou\u001b[0m: ")
			if !scanner.Scan() {
				break
			}
			userInput := scanner.Text()
			if strings.HasPrefix(userInput, "/") {
				a.runCommand(userInput)
			}

			userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
			conversation = append(conversation, userMessage)
		}

		message, err := a.runInference(ctx, conversation)
		if err != nil {
			return err
		}
		conversation = append(conversation, message.ToParam())

		toolResults := []anthropic.ContentBlockParamUnion{}
		for _, content := range message.Content {
			switch content.Type {
			case "text":
				fmt.Printf("\u001b[93mHenk\u001b[0m: %s\n", content.Text)
			case "tool_use":
				result := a.executeTool(content.ID, content.Name, content.Input)
				toolResults = append(toolResults, result)
			}
		}
		if len(toolResults) == 0 {
			readUserInput = true
			continue
		}

		readUserInput = false
		conversation = append(conversation, anthropic.NewUserMessage(toolResults...))
	}
	return nil
}

func (a *Agent) runCommand(input string) {
	cmd, _, _ := strings.Cut(input, " ")
	cmd = strings.TrimPrefix(cmd, "/")
	switch cmd {
	case "quit":
		a.done <- true
	}

}

func (a *Agent) executeTool(id, name string, input json.RawMessage) anthropic.ContentBlockParamUnion {
	var t tool.Tool
	var found bool
	for _, i := range a.tools {
		if i.Name() == name {
			t = i
			found = true
			break
		}
	}
	if !found {
		return anthropic.NewToolResultBlock(id, "tool not found", true)
	}
	fmt.Printf("\u001b[92mtool\u001b[0m: %s(%s)\n", name, input)
	response, err := t.Execute(input)
	if err != nil {
		return anthropic.NewToolResultBlock(id, err.Error(), true)
	}

	return anthropic.NewToolResultBlock(id, response, false)
}

func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	anthropicTools := []anthropic.ToolUnionParam{}
	for _, tool := range a.tools {
		anthropicTools = append(anthropicTools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name(),
				Description: anthropic.String(tool.Description()),
				InputSchema: tool.InputSchema(),
			},
		})
	}
	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_7SonnetLatest,
		MaxTokens: int64(1024),
		Messages:  conversation,
		Tools:     anthropicTools,
	})

	return message, err
}
