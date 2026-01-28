// Package id provides unique identifier generation for snippets.
package id

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	// Length is the number of characters in a generated ID.
	Length = 12

	// Alphabet is base62: 0-9, A-Z, a-z for URL-safe IDs.
	Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

// Generator creates unique snippet IDs.
type Generator struct{}

// New creates a new ID generator.
func New() *Generator {
	return &Generator{}
}

// Generate creates a new unique ID.
// Returns a 12-character base62 string with ~71 bits of entropy.
func (g *Generator) Generate() (string, error) {
	return gonanoid.Generate(Alphabet, Length)
}

// MustGenerate creates a new unique ID, panicking on error.
// Use only in contexts where ID generation failure is unrecoverable.
func (g *Generator) MustGenerate() string {
	id, err := g.Generate()
	if err != nil {
		panic(err)
	}
	return id
}

// IsValid checks if a string is a valid snippet ID.
func IsValid(id string) bool {
	if len(id) != Length {
		return false
	}
	for _, c := range id {
		if !isBase62(c) {
			return false
		}
	}
	return true
}

func isBase62(c rune) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z')
}
