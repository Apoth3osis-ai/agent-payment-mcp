package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
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
		if err := http.ListenAndServe(addr, srv.Router()); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait a moment for server to start
	time.Sleep(500 * time.Millisecond)

	// Open browser
	url := fmt.Sprintf("http://localhost:%s", port)
	if err := openBrowser(url); err != nil {
		log.Printf("Please open your browser to: %s", url)
	}

	// Keep server running
	log.Println("Installer running. Close this window when installation is complete.")
	select {}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // linux, freebsd, openbsd, netbsd
		cmd = exec.Command("xdg-open", url)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}
