package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://test.api.com", "api-key", "budget-key")

	if client.baseURL != "https://test.api.com" {
		t.Errorf("Expected baseURL to be https://test.api.com, got %s", client.baseURL)
	}
	if client.apiKey != "api-key" {
		t.Errorf("Expected apiKey to be api-key, got %s", client.apiKey)
	}
	if client.budgetKey != "budget-key" {
		t.Errorf("Expected budgetKey to be budget-key, got %s", client.budgetKey)
	}
}

func TestNewClientDefaultURL(t *testing.T) {
	client := NewClient("", "api-key", "budget-key")

	if client.baseURL != "https://api.agentpmt.com" {
		t.Errorf("Expected default baseURL, got %s", client.baseURL)
	}
}

func TestFetchTools(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("User-Agent") != DefaultUA {
			t.Errorf("Expected User-Agent %s, got %s", DefaultUA, r.Header.Get("User-Agent"))
		}
		if r.Header.Get("X-API-Key") != "test-api-key" {
			t.Error("Missing or incorrect X-API-Key header")
		}
		if r.Header.Get("X-Budget-Key") != "test-budget-key" {
			t.Error("Missing or incorrect X-Budget-Key header")
		}

		// Return mock response
		resp := FetchToolsResponse{
			Success: true,
			Tools: []ToolDefinition{
				{
					Name:        "test-tool",
					Description: "A test tool",
					Parameters:  json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}}}`),
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-api-key", "test-budget-key")
	ctx := context.Background()

	tools, err := client.FetchTools(ctx)
	if err != nil {
		t.Fatalf("FetchTools() failed: %v", err)
	}

	if len(tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(tools))
	}

	if tools[0].Name != "test-tool" {
		t.Errorf("Expected tool name 'test-tool', got %s", tools[0].Name)
	}

	// Verify raw JSON schema is preserved
	var schema map[string]interface{}
	if err := json.Unmarshal(tools[0].Parameters, &schema); err != nil {
		t.Errorf("Failed to parse parameters as JSON: %v", err)
	}
	if schema["type"] != "object" {
		t.Error("Schema not preserved correctly")
	}
}

func TestFetchToolsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := FetchToolsResponse{
			Success: false,
			Error:   "API error occurred",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", "test-budget")
	ctx := context.Background()

	_, err := client.FetchTools(ctx)
	if err == nil {
		t.Error("Expected error when API returns success=false")
	}
}

func TestFetchToolsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", "test-budget")
	ctx := context.Background()

	_, err := client.FetchTools(ctx)
	if err == nil {
		t.Error("Expected error when HTTP status is 500")
	}
}

func TestPurchase(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method and headers
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type: application/json")
		}

		// Verify request body
		var req PurchaseRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}
		if req.ProductID != "test-product" {
			t.Errorf("Expected product_id 'test-product', got %s", req.ProductID)
		}

		// Return success response
		resp := PurchaseResponse{
			Success: true,
			Output:  "Purchase successful",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", "test-budget")
	ctx := context.Background()

	req := PurchaseRequest{
		ProductID:  "test-product",
		Parameters: json.RawMessage(`{"param":"value"}`),
	}

	resp, err := client.Purchase(ctx, req)
	if err != nil {
		t.Fatalf("Purchase() failed: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success=true")
	}
	if resp.Output != "Purchase successful" {
		t.Errorf("Expected output 'Purchase successful', got %s", resp.Output)
	}
}

func TestPurchaseTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		json.NewEncoder(w).Encode(PurchaseResponse{Success: true})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", "test-budget")
	client.http.Timeout = 100 * time.Millisecond // Very short timeout

	ctx := context.Background()
	req := PurchaseRequest{
		ProductID:  "test",
		Parameters: json.RawMessage(`{}`),
	}

	_, err := client.Purchase(ctx, req)
	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		json.NewEncoder(w).Encode(FetchToolsResponse{Success: true})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", "test-budget")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.FetchTools(ctx)
	if err == nil {
		t.Error("Expected context cancellation error")
	}
}
