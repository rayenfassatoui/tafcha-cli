package id

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_Generate(t *testing.T) {
	gen := New()

	id, err := gen.Generate()
	require.NoError(t, err)

	assert.Len(t, id, Length)
	assert.True(t, IsValid(id), "generated ID should be valid")
}

func TestGenerator_Generate_Uniqueness(t *testing.T) {
	gen := New()
	seen := make(map[string]bool)

	// Generate 1000 IDs and ensure no duplicates
	for i := 0; i < 1000; i++ {
		id, err := gen.Generate()
		require.NoError(t, err)
		assert.False(t, seen[id], "duplicate ID generated: %s", id)
		seen[id] = true
	}
}

func TestGenerator_MustGenerate(t *testing.T) {
	gen := New()

	// Should not panic
	id := gen.MustGenerate()
	assert.Len(t, id, Length)
	assert.True(t, IsValid(id))
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		valid bool
	}{
		{"valid 12 char base62", "abc123XYZ789", true},
		{"valid all digits", "123456789012", true},
		{"valid all lowercase", "abcdefghijkl", true},
		{"valid all uppercase", "ABCDEFGHIJKL", true},
		{"too short", "abc123", false},
		{"too long", "abc123XYZ7890", false},
		{"empty", "", false},
		{"contains underscore", "abc_23XYZ789", false},
		{"contains dash", "abc-23XYZ789", false},
		{"contains space", "abc 23XYZ789", false},
		{"contains special char", "abc!23XYZ789", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, IsValid(tt.id))
		})
	}
}

func BenchmarkGenerate(b *testing.B) {
	gen := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate()
	}
}
