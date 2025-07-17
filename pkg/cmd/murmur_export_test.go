package cmd

import (
	"os"
	"strings"
	"testing"
)

// TestParseFileMode tests octal file permission parsing
func TestParseFileMode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected os.FileMode
		wantErr  bool
	}{
		{
			name:     "valid octal with leading zero",
			input:    "0600",
			expected: 0600,
			wantErr:  false,
		},
		{
			name:     "valid octal without leading zero",
			input:    "644",
			expected: 0644,
			wantErr:  false,
		},
		{
			name:     "empty string defaults to 0600",
			input:    "",
			expected: 0600,
			wantErr:  false,
		},
		{
			name:    "invalid octal number",
			input:   "888",
			wantErr: true,
		},
		{
			name:    "permissions too high",
			input:   "1000",
			wantErr: true,
		},
		{
			name:    "non-numeric input",
			input:   "abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFileMode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseFileMode(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("parseFileMode(%q) unexpected error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("parseFileMode(%q) = %o, expected %o", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseUID tests user ID parsing
func TestParseUID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		wantErr  bool
	}{
		{
			name:     "valid UID",
			input:    "1000",
			expected: 1000,
			wantErr:  false,
		},
		{
			name:     "empty string returns -1",
			input:    "",
			expected: -1,
			wantErr:  false,
		},
		{
			name:    "negative UID",
			input:   "-1",
			wantErr: true,
		},
		{
			name:    "non-numeric input",
			input:   "abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseUID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseUID(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("parseUID(%q) unexpected error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("parseUID(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestNewExportConfigFromFlags tests CLI flag validation and config creation
func TestNewExportConfigFromFlags(t *testing.T) {
	tests := []struct {
		name    string
		flags   flags
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			flags: flags{
				file:   "/tmp/secrets.env",
				format: "dotenv",
				chmod:  "0600",
				chown:  "",
			},
			wantErr: false,
		},
		{
			name: "invalid format",
			flags: flags{
				file:   "/tmp/secrets.env",
				format: "yaml",
				chmod:  "0600",
				chown:  "",
			},
			wantErr: true,
			errMsg:  "unsupported format 'yaml'",
		},
		{
			name: "empty file path",
			flags: flags{
				file:   "",
				format: "dotenv",
				chmod:  "0600",
				chown:  "",
			},
			wantErr: true,
			errMsg:  "file path cannot be empty",
		},
		{
			name: "invalid chmod",
			flags: flags{
				file:   "/tmp/secrets.env",
				format: "dotenv",
				chmod:  "invalid",
				chown:  "",
			},
			wantErr: true,
			errMsg:  "invalid chmod value",
		},
		{
			name: "invalid chown",
			flags: flags{
				file:   "/tmp/secrets.env",
				format: "dotenv",
				chmod:  "0600",
				chown:  "invalid",
			},
			wantErr: true,
			errMsg:  "invalid chown value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := newExportConfigFromFlags(&tt.flags)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewExportConfigFromFlags() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("NewExportConfigFromFlags() error = %v, expected to contain %q", err, tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("NewExportConfigFromFlags() unexpected error: %v", err)
				return
			}
			if config == nil {
				t.Errorf("NewExportConfigFromFlags() returned nil config")
			}
		})
	}
}

// TestGetEnvWithDefault tests environment variable fallback logic
func TestGetEnvWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		envVar       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "environment variable set",
			envVar:       "TEST_VAR",
			envValue:     "env_value",
			defaultValue: "default_value",
			expected:     "env_value",
		},
		{
			name:         "environment variable not set",
			envVar:       "UNSET_VAR",
			envValue:     "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.envVar)

			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			}

			result := getEnvWithDefault(tt.envVar, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvWithDefault(%q, %q) = %q, expected %q",
					tt.envVar, tt.defaultValue, result, tt.expected)
			}
		})
	}
}
