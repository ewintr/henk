package main

import (
	"fmt"
	"log"
	"os"

	"go-mod.ewintr.nl/henk/llm"
)

func main() {

	// startDir := "."
	// err := filepath.Walk(startDir, walkFunc)
	// if err != nil {
	// 	log.Fatalf("Error walking the path: %v\n", err)
	// }
	ollamaClient := llm.NewOllama("http://192.168.1.12:11434", "nomic-embed-text:latest", "qwen2.5-coder:3b-instruct-q8_0")

	response, err := ollamaClient.Complete("You are a nice person.", "Say Hi!")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(response)
}

func walkFunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.IsDir() {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Error reading file %s: %v\n", path, err)
			return nil
		}
		fmt.Printf("Contents of file %s:\n%s\n", path, string(data))
	}
	return nil
}
