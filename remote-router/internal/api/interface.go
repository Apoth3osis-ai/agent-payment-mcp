package api

import (
	"context"
)

// ClientInterface defines the methods required by MCP server
type ClientInterface interface {
	FetchTools(ctx context.Context) ([]ToolDefinition, error)
	Purchase(ctx context.Context, req PurchaseRequest) (*PurchaseResponse, error)
	StreamPurchase(ctx context.Context, req PurchaseRequest, onChunk func(string)) error
}

// Ensure Client implements ClientInterface
var _ ClientInterface = (*Client)(nil)
