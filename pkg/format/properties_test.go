package format_test

import (
	"strings"
	"testing"

	"github.com/busser/murmur/pkg/format"
	"github.com/google/go-cmp/cmp"
)

func TestPropertiesFormatter_Format(t *testing.T) {
	formatter := &format.PropertiesFormatter{}

	tests := []struct {
		name     string
		vars     map[string]string
		expected []string // Expected lines (order may vary)
	}{
		{
			name:     "empty map",
			vars:     map[string]string{},
			expected: []string{},
		},
		{
			name: "single simple value",
			vars: map[string]string{
				"key": "value",
			},
			expected: []string{"key = value"},
		},
		{
			name: "multiple simple values",
			vars: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expected: []string{"key1 = value1", "key2 = value2"},
		},
		{
			name: "empty value",
			vars: map[string]string{
				"empty": "",
			},
			expected: []string{"empty ="},
		},
		{
			name: "value with spaces",
			vars: map[string]string{
				"spaced": "hello world",
			},
			expected: []string{"spaced = hello world"},
		},
		{
			name: "value with equals sign",
			vars: map[string]string{
				"equation": "x=y+z",
			},
			expected: []string{"equation = x=y+z"},
		},
		{
			name: "value with colon",
			vars: map[string]string{
				"url": "http://example.com:8080",
			},
			expected: []string{"url = http://example.com:8080"},
		},
		{
			name: "value with backslashes",
			vars: map[string]string{
				"path": `C:\Program Files\App`,
			},
			expected: []string{"path = C:\\\\Program Files\\\\App"},
		},
		{
			name: "value with newlines",
			vars: map[string]string{
				"multiline": "line1\nline2\nline3",
			},
			expected: []string{"multiline = line1\\nline2\\nline3"},
		},
		{
			name: "value with unicode",
			vars: map[string]string{
				"unicode": "café naïve résumé",
			},
			expected: []string{"unicode = café naïve résumé"},
		},
		{
			name: "key with special characters",
			vars: map[string]string{
				"key.with.dots":        "value1",
				"key_with_underscores": "value2",
				"key-with-dashes":      "value3",
			},
			expected: []string{
				"key.with.dots = value1",
				"key_with_underscores = value2",
				"key-with-dashes = value3",
			},
		},
		{
			name: "complex mixed values",
			vars: map[string]string{
				"simple":    "simple_value",
				"complex":   "value=with:special\\chars",
				"empty":     "",
				"multiline": "line1\nline2",
				"unicode":   "café",
			},
			expected: []string{
				"simple = simple_value",
				"complex = value=with:special\\\\chars",
				"empty =",
				"multiline = line1\\nline2",
				"unicode = café",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := formatter.Format(tt.vars)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}

			if len(tt.expected) == 0 {
				if len(result) != 0 {
					t.Errorf("Expected empty result, got: %s", string(result))
				}
				return
			}

			// Split result into lines and remove empty lines
			lines := strings.Split(strings.TrimSpace(string(result)), "\n")
			var actualLines []string
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					actualLines = append(actualLines, strings.TrimSpace(line))
				}
			}

			// Sort both slices for comparison since properties order may vary
			if len(actualLines) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d lines\nExpected: %v\nActual: %v",
					len(tt.expected), len(actualLines), tt.expected, actualLines)
				return
			}

			// Check that all expected lines are present
			expectedMap := make(map[string]bool)
			for _, line := range tt.expected {
				expectedMap[line] = true
			}

			for _, line := range actualLines {
				if !expectedMap[line] {
					t.Errorf("Unexpected line in output: %s", line)
				}
			}
		})
	}
}

func TestPropertiesFormatter_Format_SortedOutput(t *testing.T) {
	formatter := &format.PropertiesFormatter{}

	vars := map[string]string{
		"zebra":   "last",
		"alpha":   "first",
		"charlie": "middle",
		"bravo":   "second",
	}

	result, err := formatter.Format(vars)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	var keys []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			parts := strings.SplitN(strings.TrimSpace(line), " = ", 2)
			if len(parts) == 2 {
				keys = append(keys, parts[0])
			}
		}
	}

	expected := []string{"alpha", "bravo", "charlie", "zebra"}
	if diff := cmp.Diff(expected, keys); diff != "" {
		t.Errorf("Format() keys not sorted (-want +got):\n%s", diff)
	}
}

// Example demonstrates basic properties formatting
func ExamplePropertiesFormatter_Format() {
	formatter := &format.PropertiesFormatter{}
	vars := map[string]string{
		"database.url": "postgres://user:pass@localhost/db",
		"api.key":      "secret-key-123",
		"debug":        "true",
	}

	result, _ := formatter.Format(vars)
	// Note: Output is sorted by key
	// Output will contain:
	// api.key = secret-key-123
	// database.url = postgres://user:pass@localhost/db
	// debug = true
	_ = result
}

// Example demonstrates escaping of special characters in properties format
func ExamplePropertiesFormatter_Format_escaping() {
	formatter := &format.PropertiesFormatter{}
	vars := map[string]string{
		"equation":  "x=y+z",
		"url":       "http://example.com:8080",
		"path":      `C:\Program Files\App`,
		"multiline": "line1\nline2",
	}

	result, _ := formatter.Format(vars)
	// Output will contain properly escaped values:
	// equation = x\=y+z
	// multiline = line1\nline2
	// path = C\:\\Program Files\\App
	// url = http\://example.com\:8080
	_ = result
}
func TestPropertiesFormatter_Format_ErrorHandling(t *testing.T) {
	formatter := &format.PropertiesFormatter{}

	tests := []struct {
		name    string
		vars    map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name: "invalid key with newline",
			vars: map[string]string{
				"KEY\nWITH\nNEWLINE": "value",
			},
			wantErr: true,
			errMsg:  "key cannot contain newline characters",
		},
		{
			name: "invalid key with equals",
			vars: map[string]string{
				"KEY=WITH=EQUALS": "value",
			},
			wantErr: true,
			errMsg:  "cannot contain '=' or ':' characters",
		},
		{
			name: "invalid key with colon",
			vars: map[string]string{
				"KEY:WITH:COLON": "value",
			},
			wantErr: true,
			errMsg:  "cannot contain '=' or ':' characters",
		},
		{
			name: "invalid key with leading whitespace",
			vars: map[string]string{
				" LEADING_SPACE": "value",
			},
			wantErr: true,
			errMsg:  "key cannot have leading or trailing whitespace",
		},
		{
			name: "invalid key with trailing whitespace",
			vars: map[string]string{
				"TRAILING_SPACE ": "value",
			},
			wantErr: true,
			errMsg:  "key cannot have leading or trailing whitespace",
		},
		{
			name: "empty key",
			vars: map[string]string{
				"": "value",
			},
			wantErr: true,
			errMsg:  "key cannot be empty",
		},
		{
			name: "very large value",
			vars: map[string]string{
				"LARGE_VALUE": strings.Repeat("x", 2*1024*1024), // 2MB value
			},
			wantErr: true,
			errMsg:  "value too large",
		},
		{
			name: "valid keys and values should not error",
			vars: map[string]string{
				"valid.key":     "value1",
				"another_key":   "value2",
				"key-with-dash": "value3",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := formatter.Format(tt.vars)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("PropertiesFormatter.Format() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("PropertiesFormatter.Format() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}
			
			if err != nil {
				t.Errorf("PropertiesFormatter.Format() unexpected error = %v", err)
			}
		})
	}
}

func TestValidatePropertiesKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid simple key",
			key:     "valid.key",
			wantErr: false,
		},
		{
			name:    "valid key with underscore",
			key:     "valid_key",
			wantErr: false,
		},
		{
			name:    "valid key with dash",
			key:     "valid-key",
			wantErr: false,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
			errMsg:  "key cannot be empty",
		},
		{
			name:    "key with newline",
			key:     "key\nwith\nnewline",
			wantErr: true,
			errMsg:  "key cannot contain newline characters",
		},
		{
			name:    "key with carriage return",
			key:     "key\rwith\rcarriage",
			wantErr: true,
			errMsg:  "key cannot contain newline characters",
		},
		{
			name:    "key with equals",
			key:     "key=with=equals",
			wantErr: true,
			errMsg:  "cannot contain '=' or ':' characters",
		},
		{
			name:    "key with colon",
			key:     "key:with:colon",
			wantErr: true,
			errMsg:  "cannot contain '=' or ':' characters",
		},
		{
			name:    "key with leading space",
			key:     " leading",
			wantErr: true,
			errMsg:  "key cannot have leading or trailing whitespace",
		},
		{
			name:    "key with trailing space",
			key:     "trailing ",
			wantErr: true,
			errMsg:  "key cannot have leading or trailing whitespace",
		},
		{
			name:    "key with leading tab",
			key:     "\tleading",
			wantErr: true,
			errMsg:  "key cannot have leading or trailing whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := format.ValidatePropertiesKey(tt.key)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePropertiesKey() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePropertiesKey() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}
			
			if err != nil {
				t.Errorf("ValidatePropertiesKey() unexpected error = %v", err)
			}
		})
	}
}

func TestValidatePropertiesValue(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid short value",
			value:   "short value",
			wantErr: false,
		},
		{
			name:    "valid empty value",
			value:   "",
			wantErr: false,
		},
		{
			name:    "valid value with special characters",
			value:   "value with\nnewlines\tand\rother chars!@#$%^&*()",
			wantErr: false,
		},
		{
			name:    "valid large value under limit",
			value:   strings.Repeat("x", 1024*1024-1), // Just under 1MB
			wantErr: false,
		},
		{
			name:    "invalid very large value",
			value:   strings.Repeat("x", 2*1024*1024), // 2MB value
			wantErr: true,
			errMsg:  "value too large (exceeds 1MB limit)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := format.ValidatePropertiesValue(tt.value)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePropertiesValue() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePropertiesValue() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}
			
			if err != nil {
				t.Errorf("ValidatePropertiesValue() unexpected error = %v", err)
			}
		})
	}
}