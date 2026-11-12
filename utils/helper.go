package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"
	"unicode"
)

// Dump serializes a value for structured logging.
func Dump(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%+v", value)
	}
	return string(raw)
}

// ExpectedString trims surrounding whitespace from user input.
func ExpectedString(value string) string {
	return strings.TrimSpace(value)
}

// ValueOrDefault returns defaultValue when value is the zero value of its type.
func ValueOrDefault[T comparable](value, defaultValue T) T {
	var zero T
	if reflect.DeepEqual(value, zero) {
		return defaultValue
	}
	return value
}

// ParseDurationWithDefault parses a duration string, falling back to a default.
func ParseDurationWithDefault(value string, defaultValue time.Duration) time.Duration {
	if value == "" {
		return defaultValue
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// WriteStringTemplate is a small alias for fmt.Sprintf used across services.
func WriteStringTemplate(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

// CalculatePages returns the number of pages for a given item count and page size.
func CalculatePages(count, size uint) int {
	if size == 0 {
		return 0
	}
	return int(math.Ceil(float64(count) / float64(size)))
}

// ToCamelCase converts snake_case/kebab-case identifiers to camelCase.
func ToCamelCase(value string) string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	if len(parts) == 0 {
		return value
	}

	var builder strings.Builder
	builder.WriteString(strings.ToLower(parts[0]))
	for _, part := range parts[1:] {
		if part == "" {
			continue
		}
		runes := []rune(strings.ToLower(part))
		runes[0] = unicode.ToUpper(runes[0])
		builder.WriteString(string(runes))
	}
	return builder.String()
}

// TruncateRunes truncates a string to max runes without breaking multibyte characters.
func TruncateRunes(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}
