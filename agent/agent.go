package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go-mod.ewintr.nl/henk/llm"
	"go-mod.ewintr.nl/henk/tool"
)

type Agent struct {
	llmClient llm.LLM
	tools     []tool.Tool
	out       chan Message
	in        chan string
	done      bool
	ctx       context.Context
}

func New(ctx context.Context, llmClient llm.LLM, tools []tool.Tool, out chan Message, in chan string) *Agent {
	return &Agent{
		llmClient: llmClient,
		tools:     tools,
		out:       out,
		in:        in,
		ctx:       ctx,
	}
}

func (a *Agent) Run() error {
	// ui sends signal when started
	<-a.in
	go a.converse()

	<-a.ctx.Done()
	return nil
}

func (a *Agent) converse() error {
	ctx := context.Background()
	conversation := make([]llm.Message, 0)

	a.out <- Message{
		Type: TypeGeneral,
		Body: "Chat with Henk (use '/quit' to quit)",
	}

	readUserInput := true
	for {
		if a.done {
			return nil
		}
		if readUserInput {
			a.out <- Message{Type: TypePrompt}
			userInput := <-a.in
			if strings.HasPrefix(userInput, "/") {
				a.runCommand(userInput)
				continue
			}

			userMessage := llm.Message{
				Role: llm.RoleUser,
				Content: []llm.ContentBlock{{
					Text: userInput,
					Type: llm.ContentTypeText,
				}},
			}
			conversation = append(conversation, userMessage)
		}

		message, err := a.llmClient.RunInference(ctx, a.tools, conversation)
		if err != nil {
			a.out <- Message{Type: TypeError, Body: err.Error()}
			continue
		}

		conversation = append(conversation, message)
		toolResults := make([]llm.Message, 0)
		for _, content := range message.Content {
			switch content.Type {
			case "text":
				a.out <- Message{Type: TypeHenk, Body: content.Text}
			case "tool_use":
				toolResult := a.executeTool(content.ToolUse.ID, content.ToolUse.Name, content.ToolUse.Input)
				toolResults = append(toolResults, llm.Message{
					Role: llm.RoleUser,
					Content: []llm.ContentBlock{{
						Type:       llm.ContentTypeToolResult,
						ToolResult: toolResult,
					}},
				})
			}
		}
		if len(toolResults) == 0 {
			readUserInput = true
			continue
		}

		readUserInput = false
		conversation = append(conversation, toolResults...)
	}
}

func (a *Agent) runCommand(input string) {
	cmd, _, _ := strings.Cut(input, " ")
	cmd = strings.TrimPrefix(cmd, "/")
	switch cmd {
	case "quit":
		a.done = true
		a.out <- Message{Type: TypeExit}
	}
}

func (a *Agent) executeTool(id, name string, input json.RawMessage) llm.ToolResult {
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
		return llm.ToolResult{
			ID:     id,
			Result: "tool not found",
			Error:  true,
		}
	}
	a.out <- Message{Type: TypeTool, Body: fmt.Sprintf("%s(%s)", name, input)}
	response, err := t.Execute(input)
	if err != nil {
		return llm.ToolResult{
			ID:     id,
			Result: err.Error(),
			Error:  true,
		}
	}

	return llm.ToolResult{
		ID:     id,
		Result: response,
	}
}
