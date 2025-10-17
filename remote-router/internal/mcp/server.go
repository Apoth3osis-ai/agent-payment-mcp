package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/Apoth3osis-ai/agent-payment-mcp/remote-router/internal/api"
)

// Server implements an MCP server over stdio
type Server struct {
	apiClient     api.ClientInterface
	version       string
	nameToIDMap   map[string]string // Maps readable name -> product ID
}

// NewServer creates a new MCP server
func NewServer(apiClient api.ClientInterface, version string) *Server {
	return &Server{
		apiClient:   apiClient,
		version:     version,
		nameToIDMap: make(map[string]string),
	}
}

// HandleStdioTransport runs the stdio transport loop
func (s *Server) HandleStdioTransport() error {
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	// Set maximum buffer size for large messages
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Bytes()

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			log.Printf("Error parsing request: %v", err)
			continue // Skip malformed requests, keep connection alive
		}

		var response JSONRPCResponse

		switch req.Method {
		case "initialize":
			response = s.handleInitialize(req.ID, req.Params)
		case "tools/list":
			response = s.handleToolsList(req.ID)
		case "tools/call":
			response = s.handleToolsCall(req.ID, req.Params)
		case "resources/list":
			// Not supported - return empty list
			response = jsonOK(req.ID, map[string]interface{}{"resources": []interface{}{}})
		case "notifications/initialized":
			// Notification - no response needed
			continue
		default:
			log.Printf("Unknown method: %s", req.Method)
			response = jsonErr(req.ID, MethodNotFound, fmt.Sprintf("method not found: %s", req.Method))
		}

		if err := encoder.Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

// handleInitialize handles the initialize method
func (s *Server) handleInitialize(id interface{}, params map[string]interface{}) JSONRPCResponse {
	log.Printf("Initialize request from client")

	return jsonOK(id, map[string]interface{}{
		"protocolVersion": ProtocolVersion,
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": true,
			},
		},
		"serverInfo": map[string]interface{}{
			"name":    "agent-payment-router",
			"version": s.version,
		},
	})
}

// extractReadableName extracts a human-readable name from the description
// and converts it to a valid MCP tool name (alphanumeric, hyphens, underscores only, max 64 chars)
func extractReadableName(description string) string {
	// Description format: "Smart Math Interpreter — A universal math engine..."
	// Extract the part before "—" or similar delimiters
	delimiters := []string{" — ", " - ", " – ", "|"}

	var readablePart string
	for _, delim := range delimiters {
		if idx := strings.Index(description, delim); idx > 0 {
			readablePart = strings.TrimSpace(description[:idx])
			break
		}
	}

	// Fallback: use first sentence or first 50 chars
	if readablePart == "" {
		if idx := strings.Index(description, "."); idx > 0 && idx < 100 {
			readablePart = strings.TrimSpace(description[:idx])
		} else if len(description) > 50 {
			readablePart = strings.TrimSpace(description[:50])
		} else {
			readablePart = strings.TrimSpace(description)
		}
	}

	// Convert to valid MCP tool name: alphanumeric, hyphens, underscores only
	// Replace spaces with hyphens
	name := strings.Join(strings.Fields(readablePart), "-")

	// Remove any characters that aren't alphanumeric, hyphen, or underscore
	var validName strings.Builder
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '-' || ch == '_' {
			validName.WriteRune(ch)
		}
	}

	result := validName.String()

	// Ensure max 64 characters (MCP requirement)
	if len(result) > 64 {
		result = result[:64]
	}

	// Trim trailing hyphens
	result = strings.TrimRight(result, "-")

	return result
}

// handleToolsList handles the tools/list method
func (s *Server) handleToolsList(id interface{}) JSONRPCResponse {
	ctx := context.Background()

	tools, err := s.apiClient.FetchTools(ctx)
	if err != nil {
		log.Printf("Failed to fetch tools: %v", err)
		return jsonErr(id, InternalError, fmt.Sprintf("failed to fetch tools: %v", err))
	}

	log.Printf("Fetched %d tools from API", len(tools))

	// Convert to MCP format with readable names and build mapping
	mcpTools := make([]MCPTool, len(tools))
	for i, tool := range tools {
		// Extract readable name from description
		readableName := extractReadableName(tool.Description)

		// Store mapping: readable name -> product ID
		s.nameToIDMap[readableName] = tool.Name

		mcpTools[i] = MCPTool{
			Name:        readableName,
			Description: tool.Description,
			InputSchema: tool.Parameters, // Raw pass-through!
		}
	}

	log.Printf("Mapped %d tools with readable names", len(s.nameToIDMap))

	return jsonOK(id, map[string]interface{}{"tools": mcpTools})
}

// handleToolsCall handles the tools/call method
func (s *Server) handleToolsCall(id interface{}, params map[string]interface{}) JSONRPCResponse {
	// Extract tool name (this will be the readable name from Claude)
	readableName, ok := params["name"].(string)
	if !ok {
		return jsonErr(id, InvalidParams, "missing or invalid 'name' parameter")
	}

	// Map readable name back to product ID
	productID, exists := s.nameToIDMap[readableName]
	if !exists {
		// Fallback: use the name as-is if not in map (shouldn't happen)
		log.Printf("Warning: Tool '%s' not found in mapping, using as-is", readableName)
		productID = readableName
	}

	// Extract arguments
	args, ok := params["arguments"].(map[string]interface{})
	if !ok {
		args = make(map[string]interface{})
	}

	log.Printf("Tool call: %s (product ID: %s)", readableName, productID)

	// Check if streaming is requested
	streaming := false
	if streamParam, ok := args["stream"].(bool); ok {
		streaming = streamParam
	}

	// Marshal arguments to JSON
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return jsonErr(id, InternalError, fmt.Sprintf("failed to marshal arguments: %v", err))
	}

	req := api.PurchaseRequest{
		ProductID:  productID, // Use the actual product ID from API
		Parameters: json.RawMessage(argsJSON),
	}

	ctx := context.Background()

	if streaming {
		// Handle streaming
		var chunks []string
		err := s.apiClient.StreamPurchase(ctx, req, func(chunk string) {
			chunks = append(chunks, chunk)
		})

		if err != nil {
			log.Printf("Streaming purchase failed: %v", err)
			return s.errorResult(id, err.Error())
		}

		output := strings.Join(chunks, "")
		log.Printf("Streaming purchase completed: %d chars", len(output))

		return s.successResult(id, output)
	}

	// Handle synchronous
	resp, err := s.apiClient.Purchase(ctx, req)
	if err != nil {
		log.Printf("Purchase failed: %v", err)
		return s.errorResult(id, err.Error())
	}

	log.Printf("Purchase completed successfully")

	return s.successResult(id, resp.Output)
}

// successResult creates a successful tool call result
func (s *Server) successResult(id interface{}, output string) JSONRPCResponse {
	return jsonOK(id, MCPToolCallResult{
		Content: []MCPContent{
			{
				Type: "text",
				Text: output,
			},
		},
		IsError: false,
	})
}

// errorResult creates an error tool call result (keeps connection alive)
func (s *Server) errorResult(id interface{}, message string) JSONRPCResponse {
	return jsonOK(id, MCPToolCallResult{
		Content: []MCPContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Error: %s", message),
			},
		},
		IsError: true,
	})
}

// SetupLogging configures logging to stderr with redaction
func SetupLogging(apiKey, budgetKey string) {
	log.SetOutput(os.Stderr)
	log.SetPrefix("[AgentPMT] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Create a custom writer that redacts secrets
	log.SetOutput(&redactingWriter{
		output:    os.Stderr,
		secrets:   []string{apiKey, budgetKey},
		prefix:    "[AgentPMT] ",
		showFlags: true,
	})
}

// redactingWriter wraps an io.Writer and redacts secrets
type redactingWriter struct {
	output    io.Writer
	secrets   []string
	prefix    string
	showFlags bool
}

func (w *redactingWriter) Write(p []byte) (n int, err error) {
	s := string(p)

	// Redact secrets
	for _, secret := range w.secrets {
		if secret != "" && len(secret) > 8 {
			redacted := secret[:4] + "***" + secret[len(secret)-4:]
			s = strings.ReplaceAll(s, secret, redacted)
		} else if secret != "" {
			s = strings.ReplaceAll(s, secret, "***")
		}
	}

	return w.output.Write([]byte(s))
}
