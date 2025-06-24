package main

import (
	"context"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"go-mod.ewintr.nl/henk/agent"
	"go-mod.ewintr.nl/henk/llm"
	"go-mod.ewintr.nl/henk/tool"
)

const (
	henkDir = ".henk"
)

func main() {
	if err := setupDir(henkDir); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	configPath := fmt.Sprintf("%s/config.toml", henkDir)
	config, err := readConfig(configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := config.Validate(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	llmClient, err := llm.NewLLM(config.Provider())
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

func setupDir(path string) error {
	info, err := os.Stat(path)
	switch {
	case os.IsNotExist(err):
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	case err != nil:
		return err
	case !info.IsDir():
		return fmt.Errorf("%s exists, but is not a directory", path)
	}

	return nil
}

func readConfig(path string) (Config, error) {
	var config Config
	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return Config{}, fmt.Errorf("could not read config file: %v", err)
	}

	return config, nil
}
