package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/Apoth3osis-ai/agent-payment-mcp/remote-router/internal/api"
)

// mockAPIClient implements a simple mock for testing
type mockAPIClient struct {
	tools []api.ToolDefinition
	purchaseResponse *api.PurchaseResponse
	purchaseError error
}

func (m *mockAPIClient) FetchTools(ctx context.Context) ([]api.ToolDefinition, error) {
	return m.tools, nil
}

func (m *mockAPIClient) Purchase(ctx context.Context, req api.PurchaseRequest) (*api.PurchaseResponse, error) {
	if m.purchaseError != nil {
		return nil, m.purchaseError
	}
	return m.purchaseResponse, nil
}

func (m *mockAPIClient) StreamPurchase(ctx context.Context, req api.PurchaseRequest, onChunk func(string)) error {
	if m.purchaseError != nil {
		return m.purchaseError
	}
	if m.purchaseResponse != nil {
		onChunk(m.purchaseResponse.Output)
	}
	return nil
}

func TestHandleInitialize(t *testing.T) {
	server := NewServer(nil, "1.0.0-test")

	resp := server.handleInitialize(1, nil)

	if resp.JSONRPC != "2.0" {
		t.Error("Expected JSONRPC 2.0")
	}
	if resp.ID != 1 {
		t.Errorf("Expected ID 1, got %v", resp.ID)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	if result["protocolVersion"] != ProtocolVersion {
		t.Errorf("Expected protocol version %s", ProtocolVersion)
	}

	serverInfo, ok := result["serverInfo"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected serverInfo")
	}

	if serverInfo["name"] != "agent-payment-router" {
		t.Error("Expected server name 'agent-payment-router'")
	}
	if serverInfo["version"] != "1.0.0-test" {
		t.Errorf("Expected version '1.0.0-test', got %s", serverInfo["version"])
	}
}

func TestHandleToolsList(t *testing.T) {
	mockClient := &mockAPIClient{
		tools: []api.ToolDefinition{
			{
				Name:        "test-tool",
				Description: "A test tool",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}}}`),
			},
		},
	}

	server := NewServer(mockClient, "1.0.0")
	resp := server.handleToolsList(2)

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	tools, ok := result["tools"].([]MCPTool)
	if !ok {
		t.Fatal("Expected tools to be []MCPTool")
	}

	if len(tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(tools))
	}

	if tools[0].Name != "test-tool" {
		t.Errorf("Expected tool name 'test-tool', got %s", tools[0].Name)
	}

	// Verify raw JSON schema preserved
	var schema map[string]interface{}
	if err := json.Unmarshal(tools[0].InputSchema, &schema); err != nil {
		t.Fatalf("Failed to parse input schema: %v", err)
	}
	if schema["type"] != "object" {
		t.Error("Schema not preserved correctly")
	}
}

func TestHandleToolsCallSuccess(t *testing.T) {
	mockClient := &mockAPIClient{
		purchaseResponse: &api.PurchaseResponse{
			Success: true,
			Output:  "Tool executed successfully",
		},
	}

	server := NewServer(mockClient, "1.0.0")

	params := map[string]interface{}{
		"name": "test-tool",
		"arguments": map[string]interface{}{
			"param1": "value1",
		},
	}

	resp := server.handleToolsCall(3, params)

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(MCPToolCallResult)
	if !ok {
		t.Fatalf("Expected MCPToolCallResult, got %T", resp.Result)
	}

	if result.IsError {
		t.Error("Expected IsError to be false")
	}

	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(result.Content))
	}

	if result.Content[0].Type != "text" {
		t.Error("Expected content type 'text'")
	}
	if result.Content[0].Text != "Tool executed successfully" {
		t.Errorf("Expected output text, got %s", result.Content[0].Text)
	}
}

func TestHandleToolsCallMissingName(t *testing.T) {
	server := NewServer(nil, "1.0.0")

	params := map[string]interface{}{
		"arguments": map[string]interface{}{},
	}

	resp := server.handleToolsCall(4, params)

	if resp.Error == nil {
		t.Error("Expected error when name is missing")
	}
	if resp.Error.Code != InvalidParams {
		t.Errorf("Expected InvalidParams error code, got %d", resp.Error.Code)
	}
}

func TestRedactingWriter(t *testing.T) {
	var buf bytes.Buffer
	writer := &redactingWriter{
		output:  &buf,
		secrets: []string{"secret-key-12345", "another-secret"},
	}

	message := "API key: secret-key-12345 and another-secret here"
	writer.Write([]byte(message))

	result := buf.String()

	if bytes.Contains([]byte(result), []byte("secret-key-12345")) {
		t.Error("Secret key was not redacted")
	}
	if bytes.Contains([]byte(result), []byte("another-secret")) {
		t.Error("Another secret was not redacted")
	}
	if !bytes.Contains([]byte(result), []byte("secr***2345")) {
		t.Error("Expected redacted format not found")
	}
}

func TestJSONRPCHelpers(t *testing.T) {
	// Test jsonOK
	resp := jsonOK(123, map[string]string{"status": "ok"})
	if resp.JSONRPC != "2.0" {
		t.Error("Expected JSONRPC 2.0")
	}
	if resp.ID != 123 {
		t.Error("ID mismatch")
	}
	if resp.Error != nil {
		t.Error("Expected no error")
	}

	// Test jsonErr
	errResp := jsonErr(456, InternalError, "Something went wrong")
	if errResp.Error == nil {
		t.Fatal("Expected error")
	}
	if errResp.Error.Code != InternalError {
		t.Errorf("Expected error code %d, got %d", InternalError, errResp.Error.Code)
	}
	if errResp.Error.Message != "Something went wrong" {
		t.Error("Error message mismatch")
	}
	if errResp.Result != nil {
		t.Error("Expected no result on error")
	}
}

func TestStdioLoop(t *testing.T) {
	mockClient := &mockAPIClient{
		tools: []api.ToolDefinition{
			{Name: "test", Description: "Test", Parameters: json.RawMessage(`{}`)},
		},
	}

	server := NewServer(mockClient, "1.0.0")

	// Create pipe for stdin/stdout simulation
	stdin := bytes.NewBufferString(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
`)

	stdout := &bytes.Buffer{}

	// Temporarily replace os.Stdin and os.Stdout
	oldStdin := io.Reader(stdin)
	oldStdout := io.Writer(stdout)

	// Read from stdin
	scanner := bufio.NewScanner(oldStdin)
	encoder := json.NewEncoder(oldStdout)

	count := 0
	for scanner.Scan() && count < 2 {
		line := scanner.Bytes()
		var req JSONRPCRequest
		json.Unmarshal(line, &req)

		var resp JSONRPCResponse
		switch req.Method {
		case "initialize":
			resp = server.handleInitialize(req.ID, req.Params)
		case "tools/list":
			resp = server.handleToolsList(req.ID)
		}

		encoder.Encode(resp)
		count++
	}

	// Verify we got responses
	output := stdout.String()
	if !bytes.Contains([]byte(output), []byte("protocolVersion")) {
		t.Error("Expected initialize response")
	}
	if !bytes.Contains([]byte(output), []byte("tools")) {
		t.Error("Expected tools/list response")
	}
}
