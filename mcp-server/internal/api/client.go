package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	BaseURL          = "https://api.agentpmt.com"
	FetchEndpoint    = "/products/fetch"
	PurchaseEndpoint = "/products/purchase"
)

// Client handles API communication with Agent Payment API
type Client struct {
	apiKey    string
	budgetKey string
	baseURL   string
	client    *http.Client
}

// NewClient creates a new API client
func NewClient(apiKey, budgetKey string) *Client {
	return &Client{
		apiKey:    apiKey,
		budgetKey: budgetKey,
		baseURL:   BaseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ToolDefinition represents a tool from the API
type ToolDefinition struct {
	Type     string       `json:"type"`
	Function FunctionDef  `json:"function"`
}

// FunctionDef represents the function definition
type FunctionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// FetchToolsResponse represents the response from /products/fetch
type FetchToolsResponse struct {
	Success bool              `json:"success"`
	Tools   []ToolDefinition  `json:"tools"`
	Error   string            `json:"error,omitempty"`
}

// PurchaseRequest represents a tool execution request
type PurchaseRequest struct {
	ProductID  string                 `json:"product_id"`
	Parameters map[string]interface{} `json:"parameters"`
}

// PurchaseResponse represents the response from /products/purchase
type PurchaseResponse struct {
	Success         bool                   `json:"success"`
	Response        PurchaseResponseData   `json:"response"`
	PurchaseResult  string                 `json:"purchase_result,omitempty"`
	PurchaseDetails interface{}            `json:"purchase_details,omitempty"`
	Error           string                 `json:"error,omitempty"`
}

// PurchaseResponseData contains the nested response data
type PurchaseResponseData struct {
	StatusCode int                    `json:"status_code"`
	Data       PurchaseDataWrapper    `json:"data"`
	Success    bool                   `json:"success"`
}

// PurchaseDataWrapper contains the output
type PurchaseDataWrapper struct {
	Success bool        `json:"success"`
	Output  interface{} `json:"output"`
}

// FetchTools retrieves available tools from the API
func (c *Client) FetchTools(page, pageSize int) (*FetchToolsResponse, error) {
	url := fmt.Sprintf("%s%s?page=%d&page_size=%d", c.baseURL, FetchEndpoint, page, pageSize)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("X-Budget-Key", c.budgetKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tools: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var toolsResp FetchToolsResponse
	if err := json.Unmarshal(body, &toolsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !toolsResp.Success {
		return nil, fmt.Errorf("API error: %s", toolsResp.Error)
	}

	return &toolsResp, nil
}

// ExecuteTool executes a tool via the purchase endpoint
func (c *Client) ExecuteTool(productID string, parameters map[string]interface{}) (*PurchaseResponse, error) {
	reqBody := PurchaseRequest{
		ProductID:  productID,
		Parameters: parameters,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, PurchaseEndpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("X-Budget-Key", c.budgetKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute tool: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var purchaseResp PurchaseResponse
	if err := json.Unmarshal(body, &purchaseResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !purchaseResp.Success {
		return nil, fmt.Errorf("tool execution failed: %s", purchaseResp.Error)
	}

	return &purchaseResp, nil
}
