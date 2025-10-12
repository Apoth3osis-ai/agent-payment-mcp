package server

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"

	"github.com/Apoth3osis-ai/agent-payment-mcp/installer/internal/detector"
	"github.com/Apoth3osis-ai/agent-payment-mcp/installer/internal/installer"
)

//go:embed all:web
var webFiles embed.FS

type Server struct{}

func New() *Server {
	return &Server{}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/detect", s.handleDetect)
	mux.HandleFunc("/api/install", s.handleInstall)

	// Serve static files
	webFS, err := fs.Sub(webFiles, "web")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", http.FileServer(http.FS(webFS)))

	return mux
}

func (s *Server) handleDetect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	d := detector.New()
	tools := d.DetectAll()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tools)
}

func (s *Server) handleInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req installer.InstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	inst := installer.New()
	progress := inst.Install(req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}
