package parse

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
func ProcessGoFile(filePath string) error {
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
