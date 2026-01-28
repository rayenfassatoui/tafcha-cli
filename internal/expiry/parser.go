// Package expiry provides duration parsing for human-friendly expiry strings.
package expiry

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var (
	// Pattern matches formats like: 10m, 12h, 3d, 1w
	durationPattern = regexp.MustCompile(`^(\d+)([mhdw])$`)

	// Unit multipliers
	unitMultipliers = map[string]time.Duration{
		"m": time.Minute,
		"h": time.Hour,
		"d": 24 * time.Hour,
		"w": 7 * 24 * time.Hour,
	}
)

// Parse converts a human-friendly duration string to time.Duration.
// Supported formats:
//   - "10m" -> 10 minutes
//   - "12h" -> 12 hours
//   - "3d"  -> 3 days
//   - "1w"  -> 1 week
//
// Returns an error for invalid formats.
func Parse(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	matches := durationPattern.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("invalid duration format: %q (expected format like 10m, 12h, 3d, 1w)", s)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid duration value: %w", err)
	}

	if value <= 0 {
		return 0, fmt.Errorf("duration value must be positive: %d", value)
	}

	unit := matches[2]
	multiplier, ok := unitMultipliers[unit]
	if !ok {
		return 0, fmt.Errorf("unknown duration unit: %s", unit)
	}

	return time.Duration(value) * multiplier, nil
}

// MustParse is like Parse but panics on error.
// Use only for known-valid constant values.
func MustParse(s string) time.Duration {
	d, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return d
}

// Validate checks if a duration is within the allowed range.
func Validate(d, min, max time.Duration) error {
	if d < min {
		return fmt.Errorf("duration %v is less than minimum %v", d, min)
	}
	if d > max {
		return fmt.Errorf("duration %v exceeds maximum %v", d, max)
	}
	return nil
}

// Format converts a duration to a human-friendly string.
// Uses the largest appropriate unit.
func Format(d time.Duration) string {
	switch {
	case d >= 7*24*time.Hour && d%(7*24*time.Hour) == 0:
		return fmt.Sprintf("%dw", d/(7*24*time.Hour))
	case d >= 24*time.Hour && d%(24*time.Hour) == 0:
		return fmt.Sprintf("%dd", d/(24*time.Hour))
	case d >= time.Hour && d%time.Hour == 0:
		return fmt.Sprintf("%dh", d/time.Hour)
	default:
		return fmt.Sprintf("%dm", d/time.Minute)
	}
}
