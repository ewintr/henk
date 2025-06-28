package tool

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
)

type ListFilesInput struct {
	Path string `json:"path" jsonschema_description:"Relative path to list files from. Use \".\" for the current working directory."`
}

type ListFiles struct {
	inputSchema *jsonschema.Schema
}

func NewListFiles() *ListFiles {
	var schema ListFilesInput
	return &ListFiles{
		inputSchema: GenerateSchema(schema),
	}
}

func (lf *ListFiles) Name() string { return "list_files" }
func (lf *ListFiles) Description() string {
	return "List files and directories at a given path. Use \".\" for the current working directory."
}
func (lf *ListFiles) InputSchema() *jsonschema.Schema {
	return lf.inputSchema
}

func (lf *ListFiles) Execute(input json.RawMessage) (string, error) {
	var listFilesInput ListFilesInput
	if err := json.Unmarshal(input, &listFilesInput); err != nil {
		return "", err
	}

	dir := "."
	if listFilesInput.Path != "" {
		dir = listFilesInput.Path
	}

	var files []string
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if relPath != "." {
			if info.IsDir() {
				files = append(files, relPath+"/")
			} else {
				files = append(files, relPath)
			}
		}

		return nil
	}); err != nil {
		return "", err
	}

	result, err := json.Marshal(files)
	if err != nil {
		return "", err
	}

	return string(result), nil
}
