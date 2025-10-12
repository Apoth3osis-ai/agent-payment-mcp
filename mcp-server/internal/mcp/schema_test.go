package mcp

import (
	"encoding/json"
	"testing"
)

func TestSanitizeJSONSchema(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Move required:true from property to required array",
			input: `{
				"type": "object",
				"properties": {
					"name": {
						"type": "string",
						"required": true,
						"description": "The name"
					},
					"age": {
						"type": "number",
						"required": false
					}
				}
			}`,
			expected: `{
				"type": "object",
				"properties": {
					"name": {
						"type": "string",
						"description": "The name"
					},
					"age": {
						"type": "number"
					}
				},
				"required": ["name"]
			}`,
		},
		{
			name: "Preserve existing required array",
			input: `{
				"type": "object",
				"properties": {
					"field1": {
						"type": "string",
						"required": true
					},
					"field2": {
						"type": "string"
					}
				},
				"required": ["field2"]
			}`,
			expected: `{
				"type": "object",
				"properties": {
					"field1": {
						"type": "string"
					},
					"field2": {
						"type": "string"
					}
				},
				"required": ["field2", "field1"]
			}`,
		},
		{
			name: "No properties - return as is",
			input: `{
				"type": "object"
			}`,
			expected: `{
				"type": "object"
			}`,
		},
		{
			name:     "Empty schema",
			input:    ``,
			expected: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeJSONSchema(json.RawMessage(tt.input))

			// Parse both to compare as objects (ignoring whitespace differences)
			if len(tt.expected) == 0 {
				if len(result) != 0 {
					t.Errorf("Expected empty result, got %s", string(result))
				}
				return
			}

			var expectedObj, resultObj map[string]interface{}
			if err := json.Unmarshal([]byte(tt.expected), &expectedObj); err != nil {
				t.Fatalf("Failed to parse expected JSON: %v", err)
			}
			if err := json.Unmarshal(result, &resultObj); err != nil {
				t.Fatalf("Failed to parse result JSON: %v", err)
			}

			// Check that properties don't have "required" field
			if props, ok := resultObj["properties"].(map[string]interface{}); ok {
				for propName, propValue := range props {
					if propMap, ok := propValue.(map[string]interface{}); ok {
						if _, hasRequired := propMap["required"]; hasRequired {
							t.Errorf("Property %s still has 'required' field", propName)
						}
					}
				}
			}

			// Check that required array exists if there were required fields
			if expectedRequired, ok := expectedObj["required"].([]interface{}); ok && len(expectedRequired) > 0 {
				resultRequired, ok := resultObj["required"].([]interface{})
				if !ok {
					t.Errorf("Expected 'required' array in result")
					return
				}

				// Convert to string sets for comparison
				expectedSet := make(map[string]bool)
				for _, v := range expectedRequired {
					if s, ok := v.(string); ok {
						expectedSet[s] = true
					}
				}

				resultSet := make(map[string]bool)
				for _, v := range resultRequired {
					if s, ok := v.(string); ok {
						resultSet[s] = true
					}
				}

				// Check all expected fields are in result
				for field := range expectedSet {
					if !resultSet[field] {
						t.Errorf("Expected field %s in required array", field)
					}
				}
			}
		})
	}
}
