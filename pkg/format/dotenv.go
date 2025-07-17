package format

import (
	"fmt"
	"sort"
	"strings"
)

// DotenvFormatter implements the Formatter interface for dotenv format (KEY=value)
type DotenvFormatter struct{}

// Format converts key-value pairs to dotenv format with proper shell escaping
func (f *DotenvFormatter) Format(vars map[string]string) ([]byte, error) {
	if len(vars) == 0 {
		return []byte{}, nil
	}

	// Validate keys for dotenv format compatibility
	for key := range vars {
		if err := ValidateDotenvKey(key); err != nil {
			return nil, fmt.Errorf("invalid key '%s' for dotenv format: %w", key, err)
		}
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(vars))
	for key := range vars {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var lines []string
	for _, key := range keys {
		value := vars[key]
		escapedValue, err := escapeShellValue(value)
		if err != nil {
			return nil, fmt.Errorf("failed to escape value for key '%s': %w", key, err)
		}
		lines = append(lines, fmt.Sprintf("%s=%s", key, escapedValue))
	}

	return []byte(strings.Join(lines, "\n") + "\n"), nil
}

// escapeShellValue properly escapes values for shell consumption
func escapeShellValue(value string) (string, error) {
	// If value is empty, return empty quotes
	if value == "" {
		return `""`, nil
	}

	// Check if value contains only safe characters (no quoting needed)
	isSafe := true
	for _, char := range value {
		// Allow alphanumeric, underscore, hyphen, dot, forward slash, colon
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_' || char == '-' ||
			char == '.' || char == '/' || char == ':') {
			isSafe = false
			break
		}
	}

	if isSafe {
		return value, nil
	}

	// Value needs quoting - escape characters that are special inside double quotes
	escaped := strings.ReplaceAll(value, `\`, `\\`)    // Backslashes first
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)   // Double quotes
	escaped = strings.ReplaceAll(escaped, `$`, `\$`)   // Dollar signs (variable expansion)
	escaped = strings.ReplaceAll(escaped, "`", "\\`")  // Backticks (command substitution)
	escaped = strings.ReplaceAll(escaped, "\n", "\\n") // Newlines
	escaped = strings.ReplaceAll(escaped, "\r", "\\r") // Carriage returns
	escaped = strings.ReplaceAll(escaped, "\t", "\\t") // Tabs

	return `"` + escaped + `"`, nil
}

// ValidateDotenvKey validates that a key is suitable for dotenv format
func ValidateDotenvKey(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	// Check first character - must be letter or underscore
	firstChar := rune(key[0])
	if !((firstChar >= 'a' && firstChar <= 'z') || (firstChar >= 'A' && firstChar <= 'Z') || firstChar == '_') {
		return fmt.Errorf("key must start with a letter or underscore")
	}

	// Check remaining characters - must be alphanumeric or underscore
	for i, char := range key {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '_') {
			return fmt.Errorf("key contains invalid character '%c' at position %d (only letters, numbers, and underscores allowed)", char, i)
		}
	}

	return nil
}