package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tmaxmax/go-sse"
)

// StreamPurchase executes a tool with SSE streaming
func (c *Client) StreamPurchase(ctx context.Context, req PurchaseRequest, onChunk func(string)) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request with stream=true query parameter
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+PurchaseEndpoint+"?stream=true", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for SSE
	httpReq.Header.Set("User-Agent", DefaultUA)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("X-API-Key", c.apiKey)
	httpReq.Header.Set("X-Budget-Key", c.budgetKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/event-stream" && contentType != "text/event-stream; charset=utf-8" {
		// Not SSE, fall back to regular response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		var purchaseResp PurchaseResponse
		if err := json.Unmarshal(body, &purchaseResp); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if !purchaseResp.Success {
			return fmt.Errorf("purchase failed: %s", purchaseResp.Error)
		}

		// Send entire output as one chunk
		onChunk(purchaseResp.Output)
		return nil
	}

	// Parse SSE stream
	for event, err := range sse.Read(resp.Body, nil) {
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("SSE read error: %w", err)
		}

		// Handle different event types
		switch event.Type {
		case "data", "": // Default event type
			if event.Data != "" {
				onChunk(event.Data)
			}
		case "error":
			return fmt.Errorf("stream error: %s", event.Data)
		case "done":
			return nil
		}
	}

	return nil
}
