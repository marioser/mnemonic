package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/marioser/mnemonic/internal/config"
	"github.com/marioser/mnemonic/internal/domains"
)

// Server is the HTTP server for hooks, admin, and sync coordination.
type Server struct {
	cfg    *config.Config
	svc    *domains.Service
	router chi.Router
	srv    *http.Server
}

// New creates a new HTTP server.
func New(cfg *config.Config, svc *domains.Service) *Server {
	s := &Server{cfg: cfg, svc: svc}
	s.router = s.buildRouter()
	return s
}

// Router returns the chi router for testing.
func (s *Server) Router() chi.Router {
	return s.router
}

func (s *Server) buildRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", s.handleHealth)
	r.Get("/status", s.handleStatus)
	r.Get("/context", s.handleContext)
	r.Post("/session/end", s.handleSessionEnd)

	return r
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	s.srv = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return s.srv.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv != nil {
		return s.srv.Shutdown(ctx)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// GET /health
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"service": "mnemonic",
		"status":  "ok",
		"version": "0.1.0",
	})
}

// GET /status
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	counts, err := s.svc.AllCounts(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	total := 0
	for _, c := range counts {
		if c > 0 {
			total += c
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"collections":    counts,
		"total_entities": total,
		"chromadb":       s.cfg.ChromaDBURL(),
		"embeddings": map[string]string{
			"model":      "all-MiniLM-L6-v2",
			"dimensions": "384",
		},
	})
}

// GET /context — returns a summary for session-start hook injection
func (s *Server) handleContext(w http.ResponseWriter, r *http.Request) {
	counts, err := s.svc.AllCounts(r.Context())
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{
			"summary": "Mnemonic KB: unable to connect to ChromaDB",
		})
		return
	}

	total := 0
	parts := []string{}
	for domain, count := range counts {
		if domain == "references" {
			continue
		}
		if count >= 0 {
			total += count
			parts = append(parts, fmt.Sprintf("%s: %d", domain, count))
		}
	}

	summary := fmt.Sprintf("Mnemonic KB connected. %d entities total.", total)
	if len(parts) > 0 {
		summary += " Domains: "
		for i, p := range parts {
			if i > 0 {
				summary += " | "
			}
			summary += p
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"summary": summary,
	})
}

// POST /session/end — called by session-stop hook
func (s *Server) handleSessionEnd(w http.ResponseWriter, r *http.Request) {
	// For now, just acknowledge. Future: save session summary to KB.
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "acknowledged",
	})
}
