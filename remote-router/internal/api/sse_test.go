package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStreamPurchaseSSE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Accept") != "text/event-stream" {
			t.Error("Expected Accept: text/event-stream header")
		}

		// Verify query parameter
		if r.URL.Query().Get("stream") != "true" {
			t.Error("Expected stream=true query parameter")
		}

		// Send SSE stream
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("Response writer doesn't support flushing")
		}

		// Send chunks
		w.Write([]byte("data: Chunk 1\n\n"))
		flusher.Flush()

		w.Write([]byte("data: Chunk 2\n\n"))
		flusher.Flush()

		w.Write([]byte("event: done\n\n"))
		flusher.Flush()
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", "test-budget")
	ctx := context.Background()

	req := PurchaseRequest{
		ProductID:  "test-product",
		Parameters: json.RawMessage(`{}`),
	}

	var chunks []string
	err := client.StreamPurchase(ctx, req, func(chunk string) {
		chunks = append(chunks, chunk)
	})

	if err != nil {
		t.Fatalf("StreamPurchase() failed: %v", err)
	}

	if len(chunks) != 2 {
		t.Errorf("Expected 2 chunks, got %d: %v", len(chunks), chunks)
	}

	if chunks[0] != "Chunk 1" {
		t.Errorf("Expected first chunk to be 'Chunk 1', got '%s'", chunks[0])
	}
	if chunks[1] != "Chunk 2" {
		t.Errorf("Expected second chunk to be 'Chunk 2', got '%s'", chunks[1])
	}
}

func TestStreamPurchaseFallbackToRegular(t *testing.T) {
	// Server returns regular JSON instead of SSE
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := PurchaseResponse{
			Success: true,
			Output:  "Regular response",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", "test-budget")
	ctx := context.Background()

	req := PurchaseRequest{
		ProductID:  "test-product",
		Parameters: json.RawMessage(`{}`),
	}

	var chunks []string
	err := client.StreamPurchase(ctx, req, func(chunk string) {
		chunks = append(chunks, chunk)
	})

	if err != nil {
		t.Fatalf("StreamPurchase() failed: %v", err)
	}

	// Should get one chunk with the full output
	if len(chunks) != 1 {
		t.Errorf("Expected 1 chunk (fallback), got %d", len(chunks))
	}
	if chunks[0] != "Regular response" {
		t.Errorf("Expected 'Regular response', got '%s'", chunks[0])
	}
}

func TestStreamPurchaseError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Write([]byte("event: error\ndata: Something went wrong\n\n"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", "test-budget")
	ctx := context.Background()

	req := PurchaseRequest{
		ProductID:  "test-product",
		Parameters: json.RawMessage(`{}`),
	}

	err := client.StreamPurchase(ctx, req, func(chunk string) {
		t.Error("Should not receive chunks on error")
	})

	if err == nil {
		t.Error("Expected error when stream sends error event")
	}
	if !strings.Contains(err.Error(), "Something went wrong") {
		t.Errorf("Expected error message in error, got: %v", err)
	}
}

func TestStreamPurchaseHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", "test-budget")
	ctx := context.Background()

	req := PurchaseRequest{
		ProductID:  "test-product",
		Parameters: json.RawMessage(`{}`),
	}

	err := client.StreamPurchase(ctx, req, func(chunk string) {
		t.Error("Should not receive chunks on HTTP error")
	})

	if err == nil {
		t.Error("Expected error when HTTP status is 400")
	}
}

func TestStreamPurchaseContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		// Send one chunk then delay
		w.Write([]byte("data: First chunk\n\n"))
		flusher.Flush()

		time.Sleep(500 * time.Millisecond)

		w.Write([]byte("data: Second chunk\n\n"))
		flusher.Flush()
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", "test-budget")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req := PurchaseRequest{
		ProductID:  "test-product",
		Parameters: json.RawMessage(`{}`),
	}

	chunks := 0
	err := client.StreamPurchase(ctx, req, func(chunk string) {
		chunks++
	})

	// Should get error due to context timeout
	if err == nil {
		t.Error("Expected context cancellation error")
	}

	// Should have received first chunk before cancellation
	if chunks != 1 {
		t.Errorf("Expected to receive 1 chunk before cancellation, got %d", chunks)
	}
}
