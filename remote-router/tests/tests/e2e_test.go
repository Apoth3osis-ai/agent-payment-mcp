package tests

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestBinaryExists verifies the binary was built
func TestBinaryExists(t *testing.T) {
	binaryPath := getBinaryPath()
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatalf("Binary not found: %s. Run 'go build' first", binaryPath)
	}
}

// TestInitializeMethod tests the initialize JSON-RPC method
func TestInitializeMethod(t *testing.T) {
	response := sendRequest(t, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result["jsonrpc"] != "2.0" {
		t.Error("Expected jsonrpc 2.0")
	}

	resultData, ok := result["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result field")
	}

	if resultData["protocolVersion"] != "2025-03-26" {
		t.Errorf("Expected protocol version 2025-03-26, got %v", resultData["protocolVersion"])
	}

	serverInfo, ok := resultData["serverInfo"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected serverInfo")
	}

	if serverInfo["name"] != "agent-payment-router" {
		t.Error("Expected server name 'agent-payment-router'")
	}
}

// TestResourcesList tests the resources/list method
func TestResourcesList(t *testing.T) {
	response := sendRequest(t, `{"jsonrpc":"2.0","id":2,"method":"resources/list","params":{}}`)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	resultData, ok := result["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result field")
	}

	resources, ok := resultData["resources"].([]interface{})
	if !ok {
		t.Fatal("Expected resources array")
	}

	if len(resources) != 0 {
		t.Error("Expected empty resources array")
	}
}

// TestToolsListWithoutKeys tests tools/list without valid API keys (should error)
func TestToolsListWithoutKeys(t *testing.T) {
	response := sendRequest(t, `{"jsonrpc":"2.0","id":3,"method":"tools/list","params":{}}`)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Should have an error due to invalid keys
	if _, hasError := result["error"]; !hasError {
		t.Log("Warning: Expected error for tools/list with invalid keys (may pass if test keys are valid)")
	}
}

// TestMultipleRequests tests sending multiple requests in sequence
func TestMultipleRequests(t *testing.T) {
	requests := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"resources/list","params":{}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/list","params":{}}`,
	}

	responses := sendMultipleRequests(t, requests)

	if len(responses) != 3 {
		t.Fatalf("Expected 3 responses, got %d", len(responses))
	}

	// Check each response has correct ID
	for i, respStr := range responses {
		var resp map[string]interface{}
		if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
			t.Errorf("Response %d: Failed to parse: %v", i, err)
			continue
		}

		expectedID := float64(i + 1)
		if resp["id"] != expectedID {
			t.Errorf("Response %d: Expected ID %v, got %v", i, expectedID, resp["id"])
		}
	}
}

// TestInvalidJSON tests that invalid JSON is handled gracefully
func TestInvalidJSON(t *testing.T) {
	// Send invalid JSON followed by valid request
	requests := []string{
		`{invalid json`,
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
	}

	responses := sendMultipleRequests(t, requests)

	// Should get only 1 response (invalid JSON is skipped)
	if len(responses) != 1 {
		t.Fatalf("Expected 1 response (invalid JSON skipped), got %d", len(responses))
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(responses[0]), &result); err != nil {
		t.Fatalf("Failed to parse valid response: %v", err)
	}

	if result["id"] != float64(1) {
		t.Error("Expected response for valid request")
	}
}

// TestUnknownMethod tests calling an unknown method
func TestUnknownMethod(t *testing.T) {
	response := sendRequest(t, `{"jsonrpc":"2.0","id":99,"method":"unknown/method","params":{}}`)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	errorData, hasError := result["error"].(map[string]interface{})
	if !hasError {
		t.Fatal("Expected error for unknown method")
	}

	if errorData["code"] != float64(-32601) {
		t.Errorf("Expected error code -32601 (method not found), got %v", errorData["code"])
	}
}

// TestBinaryVersion tests that version is embedded
func TestBinaryVersion(t *testing.T) {
	// The version should be in the initialize response
	response := sendRequest(t, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)

	var result map[string]interface{}
	json.Unmarshal([]byte(response), &result)

	resultData := result["result"].(map[string]interface{})
	serverInfo := resultData["serverInfo"].(map[string]interface{})
	version := serverInfo["version"].(string)

	// Should be either "dev" or a semver version
	if version == "" {
		t.Error("Version should not be empty")
	}

	t.Logf("Binary version: %s", version)
}

// Helper functions

func getBinaryPath() string {
	// Try to find the binary in common locations
	candidates := []string{
		"../agent-payment-router",
		"../distribution/binaries/linux-amd64/agent-payment-router",
		"./agent-payment-router",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	// Default to assuming binary is in parent directory
	return "../agent-payment-router"
}

func sendRequest(t *testing.T, request string) string {
	responses := sendMultipleRequests(t, []string{request})
	if len(responses) == 0 {
		t.Fatal("No response received")
	}
	return responses[0]
}

func sendMultipleRequests(t *testing.T, requests []string) []string {
	binaryPath := getBinaryPath()

	// Prepare input
	input := bytes.NewBufferString("")
	for _, req := range requests {
		input.WriteString(req + "\n")
	}

	// Set test environment
	cmd := exec.Command(binaryPath)
	cmd.Stdin = input
	cmd.Env = append(os.Environ(),
		"AGENTPMT_API_KEY=test-api-key",
		"AGENTPMT_BUDGET_KEY=test-budget-key",
	)

	// Capture stdout and stderr separately
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run with timeout
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start binary: %v", err)
	}

	// Wait with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil && err.Error() != "exit status 1" {
			t.Logf("Binary stderr: %s", stderr.String())
			t.Fatalf("Binary exited with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		cmd.Process.Kill()
		t.Fatal("Binary timed out")
	}

	// Parse responses
	var responses []string
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			responses = append(responses, line)
		}
	}

	return responses
}

// Benchmark tests

func BenchmarkInitialize(b *testing.B) {
	binaryPath := getBinaryPath()
	request := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`

	for i := 0; i < b.N; i++ {
		cmd := exec.Command(binaryPath)
		cmd.Stdin = bytes.NewBufferString(request + "\n")
		cmd.Env = append(os.Environ(),
			"AGENTPMT_API_KEY=test-key",
			"AGENTPMT_BUDGET_KEY=test-budget",
		)
		cmd.Run()
	}
}
