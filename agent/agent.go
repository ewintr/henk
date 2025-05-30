package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"go-mod.ewintr.nl/henk/tool"
)

type Agent struct {
	llm   LLM
	tools []tool.Tool
	done  chan bool
}

func New(llm LLM, tools []tool.Tool) *Agent {
	return &Agent{
		llm:   llm,
		tools: tools,
		done:  make(chan bool),
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
	conversation := make([]Message, 0)
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

			userMessage := Message{
				Role: RoleUser,
				Content: []ContentBlock{{
					Text: userInput,
					Type: ContentTypeText,
				}},
			}
			conversation = append(conversation, userMessage)
		}

		message, err := a.llm.RunInference(ctx, a.tools, conversation)
		if err != nil {
			return err
		}
		conversation = append(conversation, message)

		toolResults := make([]Message, 0)
		for _, content := range message.Content {
			switch content.Type {
			case "text":
				fmt.Printf("\u001b[93mHenk\u001b[0m: %s\n", content.Text)
			case "tool_use":
				toolResult := a.executeTool(content.ToolUse.ID, content.ToolUse.Name, content.ToolUse.Input)
				toolResults = append(toolResults, Message{
					Role: RoleUser,
					Content: []ContentBlock{{
						Type:       ContentTypeToolResult,
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

func (a *Agent) executeTool(id, name string, input json.RawMessage) ToolResult {
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
		return ToolResult{
			ID:     id,
			Result: "tool not found",
			Error:  true,
		}
	}
	fmt.Printf("\u001b[92mtool\u001b[0m: %s(%s): %s\n", name, id, input)
	response, err := t.Execute(input)
	if err != nil {
		return ToolResult{
			ID:     id,
			Result: err.Error(),
			Error:  true,
		}
	}

	return ToolResult{
		ID:     id,
		Result: response,
	}
}
