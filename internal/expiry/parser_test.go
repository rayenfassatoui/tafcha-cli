package expiry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"10m", 10 * time.Minute},
		{"30m", 30 * time.Minute},
		{"1h", 1 * time.Hour},
		{"12h", 12 * time.Hour},
		{"24h", 24 * time.Hour},
		{"1d", 24 * time.Hour},
		{"3d", 72 * time.Hour},
		{"7d", 7 * 24 * time.Hour},
		{"1w", 7 * 24 * time.Hour},
		{"2w", 14 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParse_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"no unit", "10"},
		{"no value", "m"},
		{"invalid unit", "10x"},
		{"negative value", "-5m"},
		{"decimal value", "1.5h"},
		{"spaces", "10 m"},
		{"mixed case", "10M"},
		{"zero value", "0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			assert.Error(t, err)
		})
	}
}

func TestMustParse(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		result := MustParse("3d")
		assert.Equal(t, 72*time.Hour, result)
	})

	t.Run("panics on invalid input", func(t *testing.T) {
		assert.Panics(t, func() {
			MustParse("invalid")
		})
	})
}

func TestValidate(t *testing.T) {
	min := 10 * time.Minute
	max := 30 * 24 * time.Hour

	tests := []struct {
		name    string
		d       time.Duration
		wantErr bool
	}{
		{"within range", 24 * time.Hour, false},
		{"at minimum", min, false},
		{"at maximum", max, false},
		{"below minimum", 5 * time.Minute, true},
		{"above maximum", 31 * 24 * time.Hour, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.d, min, max)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{10 * time.Minute, "10m"},
		{90 * time.Minute, "90m"},
		{1 * time.Hour, "1h"},
		{12 * time.Hour, "12h"},
		{24 * time.Hour, "1d"},
		{72 * time.Hour, "3d"},
		{7 * 24 * time.Hour, "1w"},
		{14 * 24 * time.Hour, "2w"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := Format(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
