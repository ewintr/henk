package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

// printNode prints a single AST node back to Go source code.
func printNode(node ast.Node) (string, error) {
	var writer bytes.Buffer
	err := format.Node(&writer, token.NewFileSet(), node)
	if err != nil {
		return "", err
	}
	return writer.String(), nil
}

// walkFile walks through the AST and collects top-level declarations.
func walkFile(f *ast.File) ([]ast.Decl, error) {
	var topLevelDecls []ast.Decl

	for _, decl := range f.Decls {
		topLevelDecls = append(topLevelDecls, decl)
	}
	return topLevelDecls, nil
}

// processGoFile processes a Go file and prints each top-level declaration.
func processGoFile(filePath string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return fmt.Errorf("error parsing %s: %w", filePath, err)
	}

	topLevelDecls, err := walkFile(f)
	if err != nil {
		return fmt.Errorf("error walking file: %w", err)
	}

	for i, decl := range topLevelDecls {
		snippet, err := printNode(decl)
		if err != nil {
			return fmt.Errorf("error printing node: %w", err)
		}
		fmt.Printf("Top-level Declaration %d:\n%s\n---\n", i+1, snippet)
	}

	return nil
}

func main() {
	filePath := "llm/ollama.go" // Replace with your Go file path

	err := processGoFile(filePath)
	if err != nil {
		fmt.Println(err)
	}
}

// startDir := "."
// err := filepath.Walk(startDir, walkFunc)
// if err != nil {
// 	log.Fatalf("Error walking the path: %v\n", err)
// }
// ollamaClient := llm.NewOllama("http://192.168.1.12:11434", "nomic-embed-text:latest", "qwen2.5-coder:32b-instruct-q8_0")

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
