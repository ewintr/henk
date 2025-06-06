package tool

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
	"go-mod.ewintr.nl/henk/storage"
)

type FileSummaryInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a file in the working directory."`
}

type FileSummary struct {
	fileRepo    storage.FileIndex
	inputSchema *jsonschema.Schema
}

func NewFileSummary(fileRepo storage.FileIndex) *FileSummary {
	var schema FileSummaryInput
	return &FileSummary{
		fileRepo:    fileRepo,
		inputSchema: GenerateSchema(schema),
	}
}

func (fs *FileSummary) Name() string { return "file_summary" }
func (fs *FileSummary) Description() string {
	return "Fetch a summary of the file contents. Use this if you want to understand on a high level what the content of a file is, but don't need any specific details yet. Do not use this with directory names."
}
func (fs *FileSummary) InputSchema() *jsonschema.Schema {
	return fs.inputSchema
}

func (fs *FileSummary) Execute(input json.RawMessage) (string, error) {
	var fsInput FileSummaryInput
	if err := json.Unmarshal(input, &fsInput); err != nil {
		return "", err
	}

	file, err := fs.fileRepo.FindByPath(fsInput.Path)
	if err != nil {
		return "", nil
	}

	return file.Summary, nil
}
