package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// HandleStdioTransport handles JSON-RPC over stdio with custom tools/list
func (s *Server) HandleStdioTransport() error {
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		line := scanner.Bytes()

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			log.Printf("Error parsing request: %v", err)
			continue
		}

		// Handle tools/list ourselves to preserve raw schemas
		if req.Method == "tools/list" {
			response := s.handleToolsList(req.ID)
			if err := encoder.Encode(response); err != nil {
				log.Printf("Error encoding response: %v", err)
			}
			continue
		}

		// Handle initialize
		if req.Method == "initialize" {
			response := s.handleInitialize(req.ID)
			if err := encoder.Encode(response); err != nil {
				log.Printf("Error encoding response: %v", err)
			}
			continue
		}

		// Handle tools/call
		if req.Method == "tools/call" {
			response := s.handleToolsCall(req.ID, req.Params)
			if err := encoder.Encode(response); err != nil {
				log.Printf("Error encoding response: %v", err)
			}
			continue
		}

		// Handle notifications/initialized (no response needed)
		if req.Method == "notifications/initialized" {
			continue
		}

		// Handle other list methods
		if req.Method == "prompts/list" {
			response := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  map[string]interface{}{"prompts": []interface{}{}},
			}
			encoder.Encode(response)
			continue
		}

		if req.Method == "resources/list" {
			response := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  map[string]interface{}{"resources": []interface{}{}},
			}
			encoder.Encode(response)
			continue
		}

		// Unknown method
		response := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: map[string]interface{}{
				"code":    -32601,
				"message": fmt.Sprintf("Method not implemented: %s", req.Method),
			},
		}
		encoder.Encode(response)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(id interface{}) JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"protocolVersion": "2025-03-26",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": true,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "agent-payment",
				"version": "1.0.0",
			},
		},
	}
}

// handleToolsList returns tools with raw schemas preserved
func (s *Server) handleToolsList(id interface{}) JSONRPCResponse {
	s.toolsMux.RLock()
	defer s.toolsMux.RUnlock()

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"tools": s.rawTools,
		},
	}
}

// handleToolsCall executes a tool
func (s *Server) handleToolsCall(id interface{}, params json.RawMessage) JSONRPCResponse {
	var callParams struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(params, &callParams); err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: map[string]interface{}{
				"code":    -32602,
				"message": fmt.Sprintf("Invalid params: %v", err),
			},
		}
	}

	log.Printf("Executing tool: %s with arguments: %v", callParams.Name, callParams.Arguments)

	// Map display name to product ID
	s.toolsMux.RLock()
	productID, exists := s.nameToID[callParams.Name]
	s.toolsMux.RUnlock()

	if !exists {
		// If not found, assume it's already a product ID
		productID = callParams.Name
	}

	log.Printf("Mapped tool name '%s' to product ID '%s'", callParams.Name, productID)

	// Execute via API client
	result, err := s.apiClient.ExecuteTool(productID, callParams.Arguments)
	if err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": fmt.Sprintf("Tool execution failed: %v", err),
					},
				},
				"isError": true,
			},
		}
	}

	// Extract the actual output from the nested response structure
	output := result.Response.Data.Output

	// Format result as JSON for better readability
	outputJSON, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		outputJSON = []byte(fmt.Sprintf("%v", output))
	}

	resultText := fmt.Sprintf("Tool Result:\n%s", string(outputJSON))

	// Add purchase info if available
	if result.PurchaseResult != "" {
		resultText += fmt.Sprintf("\n\nPurchase: %s", result.PurchaseResult)
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": resultText,
				},
			},
		},
	}
}
