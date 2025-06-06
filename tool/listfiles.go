package tool

import (
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
	"go-mod.ewintr.nl/henk/storage"
)

type ListFilesInput struct {
	Path string `json:"path,omitempty" jsonschema_description:"Optional relative path to list files from. Defaults to current directory if not provided."`
}

type ListFiles struct {
	inputSchema anthropic.ToolInputSchemaParam
	fileRepo    storage.FileIndex
}

func NewListFiles(fileRepo storage.FileIndex) *ListFiles {
	var schema ListFilesInput
	return &ListFiles{
		inputSchema: GenerateSchema(schema),
		fileRepo:    fileRepo,
	}
}

func (lf *ListFiles) Name() string { return "list_files" }
func (lf *ListFiles) Description() string {
	return "List files and directories at a given path. If no path is provided, lists files in the current directory."
}
func (lf *ListFiles) InputSchema() anthropic.ToolInputSchemaParam {
	return lf.inputSchema
}

func (lf *ListFiles) Execute(input json.RawMessage) (string, error) {
	list, err := lf.fileRepo.ListPaths()
	if err != nil {
		return "", err
	}
	result, err := json.Marshal(list)
	if err != nil {
		return "", err
	}

	return string(result), nil
}
