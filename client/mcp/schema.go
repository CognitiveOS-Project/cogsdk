package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// schemaURL is the stable resource URL used to register tool input schemas
// with the JSON Schema compiler. It is not fetched — the schema document is
// provided in-process (shift-left validation, per ADR-011).
const schemaURL = "https://cognitive-os.org/schemas/tool-input-schema.json"

// ValidateArgs validates arguments against a JSON Schema inputSchema.
//
// A nil schema returns nil (no constraint). Invalid schema documents return an
// error before argument validation.
func ValidateArgs(inputSchema interface{}, args map[string]interface{}) error {
	if inputSchema == nil {
		return nil
	}

	doc, ok := inputSchema.(map[string]interface{})
	if !ok {
		// Accept a marshaled JSON object too.
		data, err := json.Marshal(inputSchema)
		if err != nil {
			return fmt.Errorf("invalid inputSchema: %w", err)
		}
		if err := json.Unmarshal(data, &doc); err != nil || doc == nil {
			return fmt.Errorf("invalid inputSchema: expected JSON object")
		}
	}

	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat() // return validation errors for bad format values
	if err := compiler.AddResource(schemaURL, doc); err != nil {
		return fmt.Errorf("invalid inputSchema: %w", err)
	}

	sch, err := compiler.Compile(schemaURL)
	if err != nil {
		return fmt.Errorf("compile inputSchema: %w", err)
	}

	if err := sch.Validate(args); err != nil {
		return fmt.Errorf("argument validation failed: %w", err)
	}
	return nil
}
