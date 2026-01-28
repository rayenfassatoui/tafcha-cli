package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"

	"github.com/rayenfassatoui/tafcha-cli/internal/config"
	"github.com/rayenfassatoui/tafcha-cli/internal/id"
	"github.com/rayenfassatoui/tafcha-cli/internal/storage"
)

// Server represents the HTTP API server.
type Server struct {
	router      *chi.Mux
	config      *config.Config
	repo        storage.Repository
	idGenerator *id.Generator
	logger      *slog.Logger
}

// NewServer creates a new API server.
func NewServer(cfg *config.Config, repo storage.Repository, logger *slog.Logger) *Server {
	s := &Server{
		router:      chi.NewRouter(),
		config:      cfg,
		repo:        repo,
		idGenerator: id.New(),
		logger:      logger,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	// Request ID for tracing
	s.router.Use(middleware.RequestID)

	// Real IP extraction (for rate limiting behind proxies)
	s.router.Use(middleware.RealIP)

	// Structured logging
	s.router.Use(s.loggingMiddleware)

	// Panic recovery
	s.router.Use(middleware.Recoverer)

	// Content-Type enforcement for POST
	s.router.Use(s.contentTypeMiddleware)
}

func (s *Server) setupRoutes() {
	// Health checks (no rate limiting)
	s.router.Get("/healthz", s.handleHealthz)
	s.router.Get("/readyz", s.handleReadyz)

	// POST endpoint with rate limiting
	s.router.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(s.config.PostRateLimit, time.Minute))
		r.Post("/", s.handleCreate)
	})

	// GET endpoint with rate limiting
	s.router.Group(func(r chi.Router) {
		r.Use(httprate.LimitByIP(s.config.GetRateLimit, time.Minute))
		r.Get("/{id}", s.handleGet)
	})
}

// loggingMiddleware logs HTTP requests.
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			s.logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", middleware.GetReqID(r.Context()),
				"remote_ip", r.RemoteAddr,
			)
		}()

		next.ServeHTTP(ww, r)
	})
}

// contentTypeMiddleware ensures POST requests have appropriate content type.
func (s *Server) contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only check POST requests
		if r.Method == http.MethodPost {
			ct := r.Header.Get("Content-Type")
			// Allow text/plain, application/octet-stream, or empty (defaults to text/plain)
			if ct != "" && ct != "text/plain" && ct != "application/octet-stream" {
				// Still allow it but log a warning
				s.logger.Warn("unusual content-type for POST",
					"content_type", ct,
					"request_id", middleware.GetReqID(r.Context()),
				)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// Handler returns the HTTP handler.
func (s *Server) Handler() http.Handler {
	return s.router
}
