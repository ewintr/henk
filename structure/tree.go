package structure

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PathTree struct {
	Name     string
	IsDir    bool
	Children []*PathTree
}

func BuildTree(path string) (*PathTree, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	node := &PathTree{
		Name:  filepath.Base(path),
		IsDir: fileInfo.IsDir(),
	}

	if !node.IsDir {
		return node, nil
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}
		childPath := filepath.Join(path, file.Name())
		childNode, err := BuildTree(childPath)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, childNode)
	}

	return node, nil
}

func PrintTree(node *PathTree, prefix string) {
	fmt.Printf("%v* %s\n", prefix, node.Name)
	newPrefix := prefix + "  "
	if len(node.Children) > 0 {
		for _, child := range node.Children {
			PrintTree(child, newPrefix+"  ")
		}
	}
}
