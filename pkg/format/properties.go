package format

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/magiconair/properties"
)

// PropertiesFormatter implements the Formatter interface for Java properties format
type PropertiesFormatter struct{}

// Format converts key-value pairs to Java properties format with proper escaping
func (f *PropertiesFormatter) Format(vars map[string]string) ([]byte, error) {
	if len(vars) == 0 {
		return []byte{}, nil
	}

	// Validate keys for properties format compatibility
	for key := range vars {
		if err := ValidatePropertiesKey(key); err != nil {
			return nil, fmt.Errorf("invalid key '%s' for properties format: %w", key, err)
		}
	}

	// Create a new properties instance
	props := properties.NewProperties()

	// Sort keys for consistent output
	keys := make([]string, 0, len(vars))
	for key := range vars {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Set properties in sorted order
	for _, key := range keys {
		value := vars[key]
		if err := ValidatePropertiesValue(value); err != nil {
			return nil, fmt.Errorf("invalid value for key '%s' in properties format: %w", key, err)
		}
		props.Set(key, value)
	}

	// Write to buffer
	var buf bytes.Buffer
	_, err := props.Write(&buf, properties.UTF8)
	if err != nil {
		return nil, fmt.Errorf("failed to write properties format: %w", err)
	}

	return buf.Bytes(), nil
}

// ValidatePropertiesKey validates that a key is suitable for properties format
func ValidatePropertiesKey(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	// Properties keys should not contain certain characters that could cause issues
	if strings.Contains(key, "\n") || strings.Contains(key, "\r") {
		return fmt.Errorf("key cannot contain newline characters")
	}

	if strings.Contains(key, "=") || strings.Contains(key, ":") {
		return fmt.Errorf("key cannot contain '=' or ':' characters (reserved for key-value separation)")
	}

	// Check for leading/trailing whitespace which can cause issues
	if strings.TrimSpace(key) != key {
		return fmt.Errorf("key cannot have leading or trailing whitespace")
	}

	return nil
}

// ValidatePropertiesValue validates that a value is suitable for properties format
func ValidatePropertiesValue(value string) error {
	// Properties format can handle most values, but we should check for extremely long values
	// that might cause memory issues
	if len(value) > 1024*1024 { // 1MB limit
		return fmt.Errorf("value too large (exceeds 1MB limit)")
	}

	return nil
}
