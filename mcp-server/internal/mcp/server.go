package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/agentpmt/agent-payment-mcp-server/internal/api"
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolWithRawSchema wraps a tool with a raw JSON schema
// to bypass the SDK's schema marshaling limitations
type ToolWithRawSchema struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// Server wraps the MCP server and API client
type Server struct {
	mcpServer    *mcp.Server
	apiClient    *api.Client
	tools        map[string]*api.ToolDefinition
	rawTools     []ToolWithRawSchema           // Store tools with raw schemas
	nameToID     map[string]string             // Map display name to product ID
	toolsMux     sync.RWMutex
}

// Config holds server configuration
type Config struct {
	APIKey    string
	BudgetKey string
}

// NewServer creates and initializes a new MCP server
func NewServer(cfg Config) (*Server, error) {
	// Create API client
	apiClient := api.NewClient(cfg.APIKey, cfg.BudgetKey)

	// Fetch tools from API
	log.Println("Fetching tools from Agent Payment API...")
	toolsResp, err := apiClient.FetchTools(1, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tools: %w", err)
	}

	log.Printf("Fetched %d tools from API", len(toolsResp.Tools))

	// Create MCP server
	mcpServer := mcp.NewServer("agent-payment", "1.0.0", nil)

	// Create server instance
	srv := &Server{
		mcpServer: mcpServer,
		apiClient: apiClient,
		tools:     make(map[string]*api.ToolDefinition),
		nameToID:  make(map[string]string),
	}

	// Register all tools dynamically
	for _, tool := range toolsResp.Tools {
		if err := srv.registerTool(tool); err != nil {
			log.Printf("Warning: failed to register tool %s: %v", tool.Function.Name, err)
			continue
		}
	}

	log.Printf("Successfully registered %d tools", len(srv.tools))

	return srv, nil
}

// registerTool registers a single tool with the MCP server
func (s *Server) registerTool(toolDef api.ToolDefinition) error {
	s.toolsMux.Lock()
	defer s.toolsMux.Unlock()

	// Store tool definition for later reference
	s.tools[toolDef.Function.Name] = &toolDef

	// Extract human-readable name from description (before "—")
	displayName := extractToolName(toolDef.Function.Description)
	cleanDescription := extractCleanDescription(toolDef.Function.Description)

	// Convert display name to MCP-compliant tool name (alphanumeric, hyphens, underscores only)
	mcpToolName := convertToMCPName(displayName)

	// Map both display name and MCP name to product ID for execution
	s.nameToID[displayName] = toolDef.Function.Name
	s.nameToID[mcpToolName] = toolDef.Function.Name

	// Sanitize schema to be JSON Schema 2020-12 compliant
	// Fixes: "required": true in properties, and default value types
	sanitizedParams := sanitizeJSONSchema(toolDef.Function.Parameters)

	// Fix sentence case in parameter descriptions/examples
	fixedParams := fixSentenceCaseInSchema(sanitizedParams)

	// Log if schema was modified during sanitization
	if string(sanitizedParams) != string(toolDef.Function.Parameters) {
		log.Printf("Sanitized schema for tool %s (fixed 'required' fields and/or default types)", toolDef.Function.Name)
	}

	// Build full description with display name prefix for better UX
	fullDescription := displayName + " — " + cleanDescription

	// Store tool with raw schema for tools/list responses
	// Use MCP-compliant version of display name
	rawTool := ToolWithRawSchema{
		Name:        mcpToolName,                    // MCP-compliant display name
		Description: fullDescription,                 // Full description with name
		InputSchema: fixedParams,
	}
	// Ensure we have valid JSON schema
	if len(rawTool.InputSchema) == 0 || string(rawTool.InputSchema) == "null" {
		rawTool.InputSchema = []byte(`{"type":"object","properties":{}}`)
	}
	s.rawTools = append(s.rawTools, rawTool)

	// Still register with SDK for tool execution (use fixed schema)
	inputSchema := convertParametersToSchema(fixedParams)
	serverTool := &mcp.ServerTool{
		Tool: &mcp.Tool{
			Name:        mcpToolName,
			Description: fullDescription,
			InputSchema: inputSchema,
		},
		Handler: s.createToolHandler(mcpToolName),
	}
	s.mcpServer.AddTools(serverTool)

	return nil
}

// extractToolName extracts the human-readable name from the description
// Description format: "Tool Name — Description text"
func extractToolName(description string) string {
	// Find the position of the em dash separator
	if idx := indexOf(description, " — "); idx != -1 {
		return trimSpace(description[:idx])
	}
	// Fallback: try regular dash
	if idx := indexOf(description, " - "); idx != -1 {
		return trimSpace(description[:idx])
	}
	// No separator found, use first 50 chars
	if len(description) > 50 {
		return description[:50]
	}
	return description
}

// extractCleanDescription removes the name prefix from description
// Returns just the description part after "—"
func extractCleanDescription(description string) string {
	// Find the position of the em dash separator
	if idx := indexOf(description, " — "); idx != -1 {
		return trimSpace(description[idx+len(" — "):])
	}
	// Fallback: try regular dash
	if idx := indexOf(description, " - "); idx != -1 {
		return trimSpace(description[idx+len(" - "):])
	}
	// No separator found, return as is
	return description
}

// convertToMCPName converts a display name to MCP-compliant format
// MCP requires: ^[a-zA-Z0-9_-]{1,64}$ (alphanumeric, hyphens, underscores only)
// Example: "Smart Math Interpreter" → "smart-math-interpreter"
func convertToMCPName(displayName string) string {
	// Convert to lowercase
	result := strings.ToLower(displayName)

	// Replace spaces with hyphens
	result = strings.ReplaceAll(result, " ", "-")

	// Remove or replace special characters
	var cleaned strings.Builder
	for _, char := range result {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' || char == '_' {
			cleaned.WriteRune(char)
		}
		// Skip any other characters
	}

	result = cleaned.String()

	// Remove duplicate hyphens
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}

	// Trim hyphens from start and end
	result = strings.Trim(result, "-")

	// Ensure length is within 64 characters
	if len(result) > 64 {
		result = result[:64]
		// Trim trailing hyphen if any
		result = strings.TrimRight(result, "-")
	}

	// If result is empty, return a default
	if len(result) == 0 {
		return "tool"
	}

	return result
}

// Helper function to find substring index
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Helper function to trim whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)

	// Trim leading spaces
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	// Trim trailing spaces
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}

// sanitizeJSONSchema sanitizes API schemas to be JSON Schema 2020-12 compliant
// Fixes:
// 1. "required": true inside properties (moves to top-level required array)
// 2. Default values with wrong type (e.g., "3600" string → 3600 integer)
// 3. Default values not in enum list (sets to first enum value)
func sanitizeJSONSchema(parametersJSON json.RawMessage) json.RawMessage {
	if len(parametersJSON) == 0 {
		return parametersJSON
	}

	// Parse the schema into a map
	var schema map[string]interface{}
	if err := json.Unmarshal(parametersJSON, &schema); err != nil {
		// If parse fails, return as-is
		return parametersJSON
	}

	// Extract properties and find fields with "required": true
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		// No properties to fix
		return parametersJSON
	}

	// Track which fields should be in the required array
	requiredFields := []string{}

	// Check existing required array
	if existingRequired, ok := schema["required"].([]interface{}); ok {
		for _, field := range existingRequired {
			if fieldStr, ok := field.(string); ok {
				requiredFields = append(requiredFields, fieldStr)
			}
		}
	}

	// Process each property
	for propName, propValue := range properties {
		propMap, ok := propValue.(map[string]interface{})
		if !ok {
			continue
		}

		// Check for "required": true inside property
		if required, exists := propMap["required"]; exists {
			// Remove the invalid "required" from property
			delete(propMap, "required")

			// If it was true, add to required array (avoid duplicates)
			if requiredBool, ok := required.(bool); ok && requiredBool {
				found := false
				for _, existing := range requiredFields {
					if existing == propName {
						found = true
						break
					}
				}
				if !found {
					requiredFields = append(requiredFields, propName)
				}
			}
		}

		// Fix default value types based on the "type" field
		if defaultValue, hasDefault := propMap["default"]; hasDefault {
			if typeStr, hasType := propMap["type"].(string); hasType {
				propMap["default"] = convertDefaultValueType(defaultValue, typeStr)
			}
		}

		// Validate enum default values - if default is not in enum, fix it
		if enumValues, hasEnum := propMap["enum"].([]interface{}); hasEnum && len(enumValues) > 0 {
			if defaultValue, hasDefault := propMap["default"]; hasDefault {
				// Check if default is in the enum list
				defaultStr := fmt.Sprintf("%v", defaultValue)
				found := false
				for _, enumVal := range enumValues {
					if fmt.Sprintf("%v", enumVal) == defaultStr {
						found = true
						break
					}
				}
				// If not found, set to first enum value
				if !found {
					propMap["default"] = enumValues[0]
					log.Printf("Fixed enum default for property (was '%v', set to '%v')", defaultStr, enumValues[0])
				}
			}
		}

		// Recursively fix nested object properties
		if propMap["type"] == "object" {
			if nestedProps, ok := propMap["properties"].(map[string]interface{}); ok {
				fixNestedProperties(nestedProps)
			}
		}

		// Fix arrays with item schemas
		if propMap["type"] == "array" {
			if items, ok := propMap["items"].(map[string]interface{}); ok {
				// Fix default value in items if present
				if itemDefault, hasDefault := items["default"]; hasDefault {
					if itemType, hasType := items["type"].(string); hasType {
						items["default"] = convertDefaultValueType(itemDefault, itemType)
					}
				}
				// Validate enum in array items
				if enumValues, hasEnum := items["enum"].([]interface{}); hasEnum && len(enumValues) > 0 {
					if defaultValue, hasDefault := items["default"]; hasDefault {
						defaultStr := fmt.Sprintf("%v", defaultValue)
						found := false
						for _, enumVal := range enumValues {
							if fmt.Sprintf("%v", enumVal) == defaultStr {
								found = true
								break
							}
						}
						if !found {
							items["default"] = enumValues[0]
						}
					}
				}
			}
		}
	}

	// Update the schema with cleaned properties and required array
	schema["properties"] = properties
	if len(requiredFields) > 0 {
		schema["required"] = requiredFields
	}

	// Marshal back to JSON
	sanitized, err := json.Marshal(schema)
	if err != nil {
		// If marshal fails, return original
		return parametersJSON
	}

	return json.RawMessage(sanitized)
}

// convertDefaultValueType converts a default value to match its declared type
func convertDefaultValueType(value interface{}, typeStr string) interface{} {
	switch typeStr {
	case "integer":
		// Convert string to integer if needed
		if strVal, ok := value.(string); ok {
			var intVal int
			fmt.Sscanf(strVal, "%d", &intVal)
			return intVal
		}
		// Already an integer or float, ensure it's int
		if floatVal, ok := value.(float64); ok {
			return int(floatVal)
		}
		return value

	case "number":
		// Convert string to float if needed
		if strVal, ok := value.(string); ok {
			var floatVal float64
			fmt.Sscanf(strVal, "%f", &floatVal)
			return floatVal
		}
		return value

	case "boolean":
		// Convert string to boolean if needed
		if strVal, ok := value.(string); ok {
			return strVal == "true" || strVal == "True" || strVal == "TRUE"
		}
		return value

	case "string":
		// Ensure it's a string
		if strVal, ok := value.(string); ok {
			return strVal
		}
		// Convert other types to string
		return fmt.Sprintf("%v", value)

	default:
		return value
	}
}

// fixNestedProperties recursively fixes nested object properties
func fixNestedProperties(properties map[string]interface{}) {
	for _, propValue := range properties {
		propMap, ok := propValue.(map[string]interface{})
		if !ok {
			continue
		}

		// Fix default value types
		if defaultValue, hasDefault := propMap["default"]; hasDefault {
			if typeStr, hasType := propMap["type"].(string); hasType {
				propMap["default"] = convertDefaultValueType(defaultValue, typeStr)
			}
		}

		// Validate enum default values
		if enumValues, hasEnum := propMap["enum"].([]interface{}); hasEnum && len(enumValues) > 0 {
			if defaultValue, hasDefault := propMap["default"]; hasDefault {
				defaultStr := fmt.Sprintf("%v", defaultValue)
				found := false
				for _, enumVal := range enumValues {
					if fmt.Sprintf("%v", enumVal) == defaultStr {
						found = true
						break
					}
				}
				if !found {
					propMap["default"] = enumValues[0]
				}
			}
		}

		// Recurse for nested objects
		if propMap["type"] == "object" {
			if nestedProps, ok := propMap["properties"].(map[string]interface{}); ok {
				fixNestedProperties(nestedProps)
			}
		}
	}
}

// fixSentenceCaseInSchema fixes sentence case in parameter descriptions
// Capitalizes first letter after "Example: " patterns, handling both escaped and unescaped quotes
func fixSentenceCaseInSchema(parametersJSON json.RawMessage) json.RawMessage {
	if len(parametersJSON) == 0 {
		return parametersJSON
	}

	// Convert to string for regex processing
	schemaStr := string(parametersJSON)

	// Pattern matches: Example: \"text (escaped quote in JSON) or Example: "text (unescaped)
	// The (?:\\"") part matches either \" or just " to handle both cases
	// We capture everything before the lowercase letter, then the lowercase letter itself
	re := regexp.MustCompile(`(Example:\s*(?:\\")?)([a-z])`)

	// Replace with capitalized version
	fixedStr := re.ReplaceAllStringFunc(schemaStr, func(match string) string {
		// Find where the lowercase letter is
		parts := re.FindStringSubmatch(match)
		if len(parts) == 3 {
			prefix := parts[1]
			letter := parts[2]
			// Capitalize the letter
			return prefix + strings.ToUpper(letter)
		}
		return match
	})

	return json.RawMessage(fixedStr)
}

// toSentenceCase converts the first letter of a string to uppercase
func toSentenceCase(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// convertParametersToSchema converts API parameter JSON to jsonschema.Schema
// Note: This is only used for SDK registration compatibility. Actual tool execution
// bypasses the SDK and uses raw JSON schemas to preserve the API's exact format.
func convertParametersToSchema(parametersJSON json.RawMessage) *jsonschema.Schema {
	// If no parameters provided, create a minimal valid schema
	if len(parametersJSON) == 0 || string(parametersJSON) == "null" {
		parametersJSON = []byte(`{"type":"object","properties":{}}`)
	}

	// Attempt to unmarshal into SDK's jsonschema.Schema
	// This may fail for non-standard schemas (e.g., "required": true inside properties)
	// but that's okay because we use raw schemas for the actual tool responses
	var schema jsonschema.Schema
	if err := json.Unmarshal(parametersJSON, &schema); err != nil {
		// Silently fall back to minimal schema - SDK registration is just for compatibility
		// Real tool execution uses raw JSON schemas via our custom RPC handler
		json.Unmarshal([]byte(`{"type":"object","properties":{}}`), &schema)
	}

	return &schema
}

// createToolHandler creates a dummy handler for SDK registration
// Actual tool execution happens via our custom RPC handler
func (s *Server) createToolHandler(toolName string) mcp.ToolHandler {
	return func(ctx context.Context, session *mcp.ServerSession, req *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
		// This handler is not used since we bypass the SDK for tool execution
		// We keep it for SDK compatibility
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "Tool execution handled by custom RPC layer",
				},
			},
		}, nil
	}
}

// Run starts the MCP server with custom stdio transport
// We use a custom handler to preserve raw JSON schemas
func (s *Server) Run(ctx context.Context) error {
	log.Println("Starting MCP server on stdio transport...")
	return s.HandleStdioTransport()
}
