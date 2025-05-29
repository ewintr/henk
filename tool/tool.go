package tool

import (
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
)

type Tool interface {
	Name() string
	Description() string
	InputSchema() anthropic.ToolInputSchemaParam
	Execute(input json.RawMessage) (string, error)
}

func GenerateSchema(t any) anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}

	schema := reflector.Reflect(t)

	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
	}
}
