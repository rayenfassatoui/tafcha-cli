// Package storage provides database operations for snippets.
package storage

import "time"

// Snippet represents a stored text snippet.
type Snippet struct {
	ID        string    `json:"id"`
	Content   []byte    `json:"-"`          // Not exposed in JSON responses
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// IsExpired checks if the snippet has expired.
func (s *Snippet) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Repository defines the interface for snippet storage operations.
type Repository interface {
	// Create stores a new snippet.
	Create(id string, content []byte, expiresAt time.Time) (*Snippet, error)

	// Get retrieves a snippet by ID. Returns nil if not found or expired.
	Get(id string) (*Snippet, error)

	// Delete removes a snippet by ID.
	Delete(id string) error

	// DeleteExpired removes all expired snippets. Returns the count of deleted snippets.
	DeleteExpired() (int64, error)

	// Close releases database connections.
	Close()
}
