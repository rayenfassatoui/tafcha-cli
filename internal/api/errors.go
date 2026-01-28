// Package api provides the HTTP API server for Tafcha.
package api

import (
	"encoding/json"
	"net/http"
)

// Error codes for API responses.
const (
	ErrCodeBadRequest     = "BAD_REQUEST"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeTooLarge       = "PAYLOAD_TOO_LARGE"
	ErrCodeRateLimited    = "RATE_LIMITED"
	ErrCodeInternalError  = "INTERNAL_ERROR"
	ErrCodeInvalidExpiry  = "INVALID_EXPIRY"
	ErrCodeEmptyContent   = "EMPTY_CONTENT"
	ErrCodeInvalidID      = "INVALID_ID"
)

// APIError represents an error response.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse is the JSON structure for error responses.
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// writeError sends a JSON error response.
func writeError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: message,
		},
	}

	json.NewEncoder(w).Encode(resp)
}

// Common error helpers

func badRequest(w http.ResponseWriter, message string) {
	writeError(w, http.StatusBadRequest, ErrCodeBadRequest, message)
}

func notFound(w http.ResponseWriter) {
	writeError(w, http.StatusNotFound, ErrCodeNotFound, "snippet not found or expired")
}

func payloadTooLarge(w http.ResponseWriter, maxSize int64) {
	writeError(w, http.StatusRequestEntityTooLarge, ErrCodeTooLarge, 
		"content exceeds maximum size")
}

func rateLimited(w http.ResponseWriter) {
	writeError(w, http.StatusTooManyRequests, ErrCodeRateLimited, 
		"rate limit exceeded, please try again later")
}

func internalError(w http.ResponseWriter) {
	writeError(w, http.StatusInternalServerError, ErrCodeInternalError, 
		"an internal error occurred")
}

func invalidExpiry(w http.ResponseWriter, message string) {
	writeError(w, http.StatusBadRequest, ErrCodeInvalidExpiry, message)
}

func emptyContent(w http.ResponseWriter) {
	writeError(w, http.StatusBadRequest, ErrCodeEmptyContent, 
		"content cannot be empty")
}

func invalidID(w http.ResponseWriter) {
	writeError(w, http.StatusBadRequest, ErrCodeInvalidID, 
		"invalid snippet ID format")
}
