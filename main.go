package main

import (
	"fmt"
	"os"

	"go-mod.ewintr.nl/henk/agent"
	"go-mod.ewintr.nl/henk/llm"
	"go-mod.ewintr.nl/henk/storage"
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
	db, err := storage.NewSqlite(fmt.Sprintf("%s/henk.db", henkDir))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fileRepo := storage.NewSqliteFile(db)
	llmClient := llm.NewClaude()

	ui := agent.NewUI()

	index := agent.NewIndex(fileRepo, llmClient, ui.In(), ui.Out())
	if err := index.Refresh(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tools := []tool.Tool{tool.NewReadFile(), tool.NewListFiles(fileRepo)}
	h := agent.New(llmClient, tools, ui.In(), ui.Out())
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
