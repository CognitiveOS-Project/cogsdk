package mcp

import "testing"

func TestValidateArgsOK(t *testing.T) {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{"type": "string"},
		},
		"required": []interface{}{"name"},
	}

	if err := ValidateArgs(schema, map[string]interface{}{"name": "photo-viewer"}); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestValidateArgsMissingRequired(t *testing.T) {
	schema := map[string]interface{}{
		"type":     "object",
		"required": []interface{}{"name"},
	}

	if err := ValidateArgs(schema, map[string]interface{}{}); err == nil {
		t.Fatal("expected validation error for missing required field")
	}
}

func TestValidateArgsNilSchema(t *testing.T) {
	if err := ValidateArgs(nil, map[string]interface{}{"x": 1}); err != nil {
		t.Fatalf("nil schema should validate anything, got %v", err)
	}
}

func TestValidateArgsInvalidSchema(t *testing.T) {
	if err := ValidateArgs(map[string]interface{}{"type": 42}, map[string]interface{}{}); err == nil {
		t.Fatal("expected error for invalid schema")
	}
}
