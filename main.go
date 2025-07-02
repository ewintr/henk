package main

import (
	"context"
	"fmt"
	"os"

	"go-mod.ewintr.nl/henk/agent"
	"go-mod.ewintr.nl/henk/agent/llm"
	"go-mod.ewintr.nl/henk/agent/tool"
)

func main() {
	config, err := agent.ReadConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := config.Validate(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	prov, ok := config.Provider(config.DefaultProvider)
	if !ok {
		fmt.Println("could not find provider %q", config.DefaultProvider)
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())
	llmClient, err := llm.NewLLM(prov, config.DefaultModel, config.SystemPrompt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ui := agent.NewUI(cancel)
	tools := []tool.Tool{tool.NewReadFile(), tool.NewListFiles()}
	h := agent.New(ctx, config, llmClient, tools, ui.In(), ui.Out())
	if err := h.Run(); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}
