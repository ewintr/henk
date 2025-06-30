package main

import (
	"context"
	"fmt"
	"os"

	"go-mod.ewintr.nl/henk/agent"
	"go-mod.ewintr.nl/henk/llm"
	"go-mod.ewintr.nl/henk/tool"
)

func main() {
	config, err := readConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := config.Validate(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	llmClient, err := llm.NewLLM(config.Provider(), config.SystemPrompt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ui := agent.NewUI(cancel)
	tools := []tool.Tool{tool.NewReadFile(), tool.NewListFiles()}
	h := agent.New(ctx, llmClient, tools, ui.In(), ui.Out())
	if err := h.Run(); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}
