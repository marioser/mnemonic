package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/marioser/mnemonic/internal/chroma"
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

	// Core endpoints
	r.Get("/health", s.handleHealth)
	r.Get("/status", s.handleStatus)
	r.Get("/context", s.handleContext)

	// Hook endpoints
	r.Get("/hook/search", s.handleHookSearch)
	r.Get("/hook/recent", s.handleHookRecent)
	r.Post("/hook/save-agent-output", s.handleHookSaveAgentOutput)
	r.Post("/hook/save-session-summary", s.handleHookSaveSessionSummary)

	// Session
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

// GET /hook/search?q=text&domain=commercial — quick search for hooks
// Returns compact results optimized for hook injection (~50 tokens)
func (s *Server) handleHookSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	domain := r.URL.Query().Get("domain")

	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "q parameter required"})
		return
	}

	if s.svc == nil {
		writeJSON(w, http.StatusOK, map[string]any{"results": []any{}, "count": 0})
		return
	}

	// Build filter
	filter := chroma.NewFilter().
		Type(r.URL.Query().Get("type")).
		Client(r.URL.Query().Get("client")).
		Build()

	results, err := s.svc.Search(r.Context(), query, domain, filter, 3)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"results": []any{}, "count": 0})
		return
	}

	// Compact format for hook injection
	type compactResult struct {
		ID         string  `json:"id"`
		Title      string  `json:"title"`
		Type       string  `json:"type"`
		Domain     string  `json:"domain"`
		Similarity float64 `json:"similarity"`
	}

	compact := make([]compactResult, 0, len(results))
	for _, r := range results {
		if r.Similarity < s.cfg.Search.MinSimilarity {
			continue
		}
		compact = append(compact, compactResult{
			ID:         r.Entity.ID,
			Title:      r.Entity.Title,
			Type:       r.Entity.Type,
			Domain:     r.Entity.Domain,
			Similarity: r.Similarity,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"results": compact,
		"count":   len(compact),
		"query":   query,
	})
}

// GET /hook/recent?domain=commercial&limit=5 — recent entities for context
func (s *Server) handleHookRecent(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "domain parameter required"})
		return
	}

	if s.svc == nil {
		writeJSON(w, http.StatusOK, map[string]any{"results": []any{}, "count": 0})
		return
	}

	entities, err := s.svc.Browse(r.Context(), domain, nil, 5, 0)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"results": []any{}, "count": 0})
		return
	}

	// Compact format
	type compactEntity struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Type  string `json:"type"`
	}

	compact := make([]compactEntity, 0, len(entities))
	for _, e := range entities {
		compact = append(compact, compactEntity{
			ID:    e.ID,
			Title: e.Title,
			Type:  e.Type,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"results": compact,
		"count":   len(compact),
		"domain":  domain,
	})
}

// POST /hook/save-agent-output — save knowledge extracted from sub-agent output
func (s *Server) handleHookSaveAgentOutput(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content   string `json:"content"`
		AgentType string `json:"agent_type"`
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if req.Content == "" {
		writeJSON(w, http.StatusOK, map[string]string{"status": "skipped", "reason": "empty content"})
		return
	}

	if s.svc == nil {
		writeJSON(w, http.StatusOK, map[string]string{"status": "skipped", "reason": "no service"})
		return
	}

	// Truncate content for summary
	summary := req.Content
	if len(summary) > 500 {
		summary = summary[:500]
	}

	entity := &domains.Entity{
		Type:    "agent_output",
		Domain:  "knowledge",
		Title:   fmt.Sprintf("Agent output: %s", req.AgentType),
		Content: req.Content,
		Summary: summary,
		Source:  "hook",
		Extra: map[string]string{
			"agent_type": req.AgentType,
			"session_id": req.SessionID,
		},
	}

	id, err := s.svc.SaveEntity(r.Context(), entity)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "saved",
		"id":     id,
	})
}

// POST /hook/save-session-summary — save session summary on stop
func (s *Server) handleHookSaveSessionSummary(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SessionID string `json:"session_id"`
		Summary   string `json:"summary"`
		Topics    string `json:"topics"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	if req.Summary == "" {
		writeJSON(w, http.StatusOK, map[string]string{"status": "skipped", "reason": "empty summary"})
		return
	}

	if s.svc == nil {
		writeJSON(w, http.StatusOK, map[string]string{"status": "skipped", "reason": "no service"})
		return
	}

	entity := &domains.Entity{
		Type:    "conversation",
		Domain:  "knowledge",
		Title:   fmt.Sprintf("Session %s", req.SessionID),
		Content: req.Summary,
		Source:  "hook",
		Extra: map[string]string{
			"session_id": req.SessionID,
			"topics":     req.Topics,
		},
	}

	id, err := s.svc.SaveEntity(r.Context(), entity)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "saved",
		"id":     id,
	})
}

// POST /session/end — called by session-stop hook
func (s *Server) handleSessionEnd(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "acknowledged",
	})
}
