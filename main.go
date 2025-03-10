package main

import (
	"fmt"
	"os"

	"go-mod.ewintr.nl/henk/llm"
	"go-mod.ewintr.nl/henk/parse"
)

func main() {
	filePath := "." // Replace with your Go file path

	project, err := parse.NewProject(filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", project.Tree())

	f, err := parse.NewFile("./llm/memory.go")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ollamaClient := llm.NewOllama("http://192.168.1.12:11434", "nomic-embed-text:latest", "qwen2.5-coder:32b-instruct-q8_0")
	short, long, err := parse.Describe(f, ollamaClient)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("short: %s\n\nlong: %s\n", short, long)

	// err := structure.ProcessGoFile(filePath)
	// if err != nil {
	// 	fmt.Println(err)
	// }
}

// startDir := "."
// err := filepath.Walk(startDir, walkFunc)
// if err != nil {
// 	log.Fatalf("Error walking the path: %v\n", err)
// }

// response, err := ollamaClient.Complete("You are a nice person.", "Say Hi!")
// if err != nil {
// 	fmt.Println("Error:", err)
// 	return
// }
// fmt.Println(response)
// prompt := fmt.Sprintf("The following is a file with Go source code. Split the code up into logical snippets. Snippets are either a function, a type, a constant or a variable. List the identifier and the line range for each snippet. Respond in JSON. \n\n Here comes the source code:\n\n```\n%s\n```", sourceDoc)
// response, err := ollamaClient.CompleteWithSnippets(systemMessage, prompt)
// if err != nil {
// 	fmt.Println("Error:", err)
// 	return
// }
// fmt.Printf("%+v\n", response)
