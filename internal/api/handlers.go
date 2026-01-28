package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/rayenfassatoui/tafcha-cli/internal/expiry"
	"github.com/rayenfassatoui/tafcha-cli/internal/id"
)

// CreateResponse is the response for successful snippet creation.
type CreateResponse struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// handleCreate handles POST / for creating new snippets.
func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())

	// Parse expiry from query parameter or use default
	expiryDuration := s.config.DefaultExpiry
	if expiryStr := r.URL.Query().Get("expiry"); expiryStr != "" {
		parsed, err := expiry.Parse(expiryStr)
		if err != nil {
			invalidExpiry(w, err.Error())
			return
		}

		if err := expiry.Validate(parsed, s.config.MinExpiry, s.config.MaxExpiry); err != nil {
			invalidExpiry(w, err.Error())
			return
		}

		expiryDuration = parsed
	}

	// Read body with size limit
	limitedReader := io.LimitReader(r.Body, s.config.MaxContentSize+1)
	content, err := io.ReadAll(limitedReader)
	if err != nil {
		s.logger.Error("failed to read request body", 
			"error", err, 
			"request_id", reqID)
		internalError(w)
		return
	}

	// Check if content exceeds limit
	if int64(len(content)) > s.config.MaxContentSize {
		payloadTooLarge(w, s.config.MaxContentSize)
		return
	}

	// Check for empty content
	if len(content) == 0 {
		emptyContent(w)
		return
	}

	// Generate unique ID
	snippetID, err := s.idGenerator.Generate()
	if err != nil {
		s.logger.Error("failed to generate ID", 
			"error", err, 
			"request_id", reqID)
		internalError(w)
		return
	}

	// Calculate expiry time
	expiresAt := time.Now().Add(expiryDuration)

	// Store snippet
	snippet, err := s.repo.Create(snippetID, content, expiresAt)
	if err != nil {
		s.logger.Error("failed to store snippet", 
			"error", err, 
			"request_id", reqID)
		internalError(w)
		return
	}

	s.logger.Info("snippet created",
		"snippet_id", snippet.ID,
		"size_bytes", len(content),
		"expires_at", snippet.ExpiresAt,
		"request_id", reqID,
	)

	// Build response
	resp := CreateResponse{
		ID:        snippet.ID,
		URL:       s.config.BaseURL + "/" + snippet.ID,
		ExpiresAt: snippet.ExpiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// handleGet handles GET /{id} for retrieving snippets.
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetReqID(r.Context())
	snippetID := chi.URLParam(r, "id")

	// Validate ID format
	if !id.IsValid(snippetID) {
		invalidID(w)
		return
	}

	// Fetch snippet
	snippet, err := s.repo.Get(snippetID)
	if err != nil {
		s.logger.Error("failed to fetch snippet", 
			"error", err, 
			"snippet_id", snippetID,
			"request_id", reqID)
		internalError(w)
		return
	}

	if snippet == nil {
		notFound(w)
		return
	}

	s.logger.Info("snippet retrieved",
		"snippet_id", snippet.ID,
		"size_bytes", len(snippet.Content),
		"request_id", reqID,
	)

	// Return raw content as text/plain
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	w.Write(snippet.Content)
}

// handleHealthz handles GET /healthz for liveness probes.
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// Pinger interface for repositories that support health checks.
type Pinger interface {
	Ping(ctx context.Context) error
}

// handleReadyz handles GET /readyz for readiness probes.
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	// Check database connectivity if repo supports Ping
	if pinger, ok := s.repo.(Pinger); ok {
		if err := pinger.Ping(r.Context()); err != nil {
			s.logger.Error("readiness check failed", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"error","message":"database unavailable"}`))
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
