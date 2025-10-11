package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

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

	// Map display name to product ID for execution
	s.nameToID[displayName] = toolDef.Function.Name

	// Store tool with raw schema for tools/list responses
	rawTool := ToolWithRawSchema{
		Name:        displayName,                    // Use readable name
		Description: cleanDescription,                // Use clean description
		InputSchema: toolDef.Function.Parameters,
	}
	// Ensure we have valid JSON schema
	if len(rawTool.InputSchema) == 0 || string(rawTool.InputSchema) == "null" {
		rawTool.InputSchema = []byte(`{"type":"object","properties":{}}`)
	}
	s.rawTools = append(s.rawTools, rawTool)

	// Still register with SDK for tool execution (even though schema is wrong)
	inputSchema := convertParametersToSchema(toolDef.Function.Parameters)
	serverTool := &mcp.ServerTool{
		Tool: &mcp.Tool{
			Name:        toolDef.Function.Name,
			Description: toolDef.Function.Description,
			InputSchema: inputSchema,
		},
		Handler: s.createToolHandler(toolDef.Function.Name),
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

// convertParametersToSchema converts API parameter JSON to jsonschema.Schema
// We unmarshal directly to preserve all fields
func convertParametersToSchema(parametersJSON json.RawMessage) *jsonschema.Schema {
	// If no parameters provided, create a minimal valid schema
	if len(parametersJSON) == 0 || string(parametersJSON) == "null" {
		parametersJSON = []byte(`{"type":"object","properties":{}}`)
	}

	// Unmarshal directly into jsonschema.Schema
	// The SDK's jsonschema.Schema can unmarshal from JSON and preserve fields
	var schema jsonschema.Schema
	if err := json.Unmarshal(parametersJSON, &schema); err != nil {
		log.Printf("Warning: failed to unmarshal schema: %v", err)
		// Return minimal schema on error
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
