package tool

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
)

type Tool interface {
	Name() string
	Description() string
	InputSchema() *jsonschema.Schema
	Execute(input json.RawMessage) (string, error)
}

func GenerateSchema(t any) *jsonschema.Schema {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}

	return reflector.Reflect(t)
}
