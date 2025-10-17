package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultUA is the User-Agent header sent with all requests
const DefaultUA = "AgentPMT-MCP/1.0"

// API endpoint paths
const (
	FetchEndpoint    = "/products/fetch"
	PurchaseEndpoint = "/products/purchase"
)

// Client handles HTTP communication with AgentPMT API
type Client struct {
	baseURL   string
	apiKey    string
	budgetKey string
	http      *http.Client
}

// NewClient creates a new API client with proper timeouts and headers
func NewClient(baseURL, apiKey, budgetKey string) *Client {
	if baseURL == "" {
		baseURL = "https://api.agentpmt.com"
	}

	return &Client{
		baseURL:   baseURL,
		apiKey:    apiKey,
		budgetKey: budgetKey,
		http: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				DisableCompression:  true, // Important for SSE streaming
				MaxIdleConns:        10,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
	}
}

// do executes an HTTP request with standard headers
func (c *Client) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", DefaultUA)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("X-Budget-Key", c.budgetKey)

	return c.http.Do(req)
}

// ToolDefinition represents a tool with raw JSON schema
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"` // Raw JSON schema
}

// APIToolWrapper wraps the tool in the API response format
type APIToolWrapper struct {
	Type     string          `json:"type"`
	Function FunctionDef     `json:"function"`
}

// FunctionDef is the function inside the API tool wrapper
type FunctionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// PaginationDetails contains pagination metadata
type PaginationDetails struct {
	ToolsOnThisPage int  `json:"tools_on_this_page"`
	TotalTools      int  `json:"total_qualified_tools"`
	PageReturned    int  `json:"page_returned"`
	PageSize        int  `json:"page_size_requested"`
	TotalPages      int  `json:"total_pages"`
	HasNextPage     bool `json:"has_next_page"`
}

// FetchToolsResponse is the response from /products/fetch
type FetchToolsResponse struct {
	Success bool              `json:"success"`
	Details PaginationDetails `json:"details"`
	Tools   []APIToolWrapper  `json:"tools"`
	Error   string            `json:"error,omitempty"`
}

// FetchTools retrieves ALL available tools from the API (handles pagination automatically)
func (c *Client) FetchTools(ctx context.Context) ([]ToolDefinition, error) {
	var allTools []ToolDefinition
	page := 1
	pageSize := 50 // Request 50 tools per page

	for {
		// Build URL with pagination
		url := fmt.Sprintf("%s%s?page=%d&page_size=%d", c.baseURL, FetchEndpoint, page, pageSize)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := c.do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode/100 != 2 {
			return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
		}

		var out FetchToolsResponse
		if err := json.Unmarshal(body, &out); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		if !out.Success {
			return nil, fmt.Errorf("API error: %s", out.Error)
		}

		// Unwrap tools from API format to our format
		for _, wrapper := range out.Tools {
			allTools = append(allTools, ToolDefinition{
				Name:        wrapper.Function.Name,
				Description: wrapper.Function.Description,
				Parameters:  wrapper.Function.Parameters,
			})
		}

		// Check if there are more pages
		if !out.Details.HasNextPage {
			break
		}

		page++
	}

	return allTools, nil
}

// PurchaseRequest is the request body for /products/purchase
type PurchaseRequest struct {
	ProductID  string          `json:"product_id"`
	Parameters json.RawMessage `json:"parameters"`
}

// PurchaseResponse is the response from /products/purchase
type PurchaseResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Purchase executes a tool synchronously
func (c *Client) Purchase(ctx context.Context, req PurchaseRequest) (*PurchaseResponse, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+PurchaseEndpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var out PurchaseResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !out.Success {
		return nil, fmt.Errorf("purchase failed: %s", out.Error)
	}

	return &out, nil
}
