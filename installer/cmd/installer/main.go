package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/Apoth3osis-ai/agent-payment-mcp/installer/internal/server"
)

func main() {
	port := "8765"

	// Create server
	srv := server.New()

	// Start server in background
	go func() {
		addr := fmt.Sprintf("localhost:%s", port)
		log.Printf("Starting installer server on http://%s", addr)
		log.Printf("If browser doesn't open, manually visit: http://%s", addr)
		if err := http.ListenAndServe(addr, srv.Router()); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait a moment for server to start
	time.Sleep(1 * time.Second)

	// Open browser
	url := fmt.Sprintf("http://localhost:%s", port)
	log.Printf("Opening browser to %s...", url)
	if err := openBrowser(url); err != nil {
		log.Printf("Could not auto-open browser: %v", err)
		log.Printf("Please manually open: %s", url)
	} else {
		log.Printf("Browser opened successfully!")
	}

	// Keep server running
	log.Println("")
	log.Println("=== Agent Payment MCP Installer Running ===")
	log.Println("Close this window when installation is complete.")
	log.Println("")
	select {}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows requires empty string before URL when using start
		cmd = exec.Command("cmd", "/c", "start", "", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // linux, freebsd, openbsd, netbsd
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}
