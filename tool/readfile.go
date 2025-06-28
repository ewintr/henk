package tool

import (
	"encoding/json"
	"os"

	"github.com/invopop/jsonschema"
)

type ReadFileInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a file in the working directory."`
}

type ReadFile struct {
	inputSchema *jsonschema.Schema
}

func NewReadFile() *ReadFile {
	var schema ReadFileInput
	return &ReadFile{
		inputSchema: GenerateSchema(schema),
	}
}

func (rf *ReadFile) Name() string { return "read_file" }
func (rf *ReadFile) Description() string {
	return "Read the contents of a given relative file path. Do not use this with directory names."
}
func (rf *ReadFile) InputSchema() *jsonschema.Schema {
	return rf.inputSchema
}

func (rf *ReadFile) Execute(input json.RawMessage) (string, error) {
	readFileInput := ReadFileInput{}
	if err := json.Unmarshal(input, &readFileInput); err != nil {
		return "", err
	}

	content, err := os.ReadFile(readFileInput.Path)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
