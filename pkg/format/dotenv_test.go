package format_test

import (
	"strings"
	"testing"

	"github.com/busser/murmur/pkg/format"
	"github.com/google/go-cmp/cmp"
)

func TestDotenvFormatter_Format(t *testing.T) {
	formatter := &format.DotenvFormatter{}

	tests := []struct {
		name     string
		vars     map[string]string
		expected string
	}{
		{
			name:     "empty map",
			vars:     map[string]string{},
			expected: "",
		},
		{
			name: "single simple value",
			vars: map[string]string{
				"KEY": "value",
			},
			expected: "KEY=value\n",
		},
		{
			name: "multiple simple values",
			vars: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
			expected: "KEY1=value1\nKEY2=value2\n",
		},
		{
			name: "empty value",
			vars: map[string]string{
				"EMPTY": "",
			},
			expected: "EMPTY=\"\"\n",
		},
		{
			name: "value with spaces",
			vars: map[string]string{
				"SPACED": "hello world",
			},
			expected: "SPACED=\"hello world\"\n",
		},
		{
			name: "value with quotes",
			vars: map[string]string{
				"QUOTED": `say "hello"`,
			},
			expected: "QUOTED=\"say \\\"hello\\\"\"\n",
		},
		{
			name: "value with backslashes",
			vars: map[string]string{
				"BACKSLASH": `path\to\file`,
			},
			expected: "BACKSLASH=\"path\\\\to\\\\file\"\n",
		},
		{
			name: "value with dollar signs",
			vars: map[string]string{
				"DOLLAR": "$HOME/path",
			},
			expected: "DOLLAR=\"\\$HOME/path\"\n",
		},
		{
			name: "value with backticks",
			vars: map[string]string{
				"BACKTICK": "`command`",
			},
			expected: "BACKTICK=\"\\`command\\`\"\n",
		},
		{
			name: "value with newlines",
			vars: map[string]string{
				"MULTILINE": "line1\nline2",
			},
			expected: "MULTILINE=\"line1\\nline2\"\n",
		},
		{
			name: "value with special shell characters",
			vars: map[string]string{
				"SPECIAL": "a|b&c;d(e)f<g>h*i?j[k]l{m}n~o#p",
			},
			expected: "SPECIAL=\"a|b&c;d(e)f<g>h*i?j[k]l{m}n~o#p\"\n",
		},
		{
			name: "complex mixed values",
			vars: map[string]string{
				"SIMPLE":    "simple",
				"COMPLEX":   `"quoted" $var with\backslash`,
				"EMPTY":     "",
				"MULTILINE": "line1\nline2\nline3",
			},
			expected: "COMPLEX=\"\\\"quoted\\\" \\$var with\\\\backslash\"\nEMPTY=\"\"\nMULTILINE=\"line1\\nline2\\nline3\"\nSIMPLE=simple\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := formatter.Format(tt.vars)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}

			if diff := cmp.Diff(tt.expected, string(result)); diff != "" {
				t.Errorf("Format() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDotenvFormatter_Format_SortedOutput(t *testing.T) {
	formatter := &format.DotenvFormatter{}

	vars := map[string]string{
		"ZEBRA":   "last",
		"ALPHA":   "first",
		"CHARLIE": "middle",
		"BRAVO":   "second",
	}

	result, err := formatter.Format(vars)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	lines := strings.Split(strings.TrimSuffix(string(result), "\n"), "\n")
	expected := []string{
		"ALPHA=first",
		"BRAVO=second",
		"CHARLIE=middle",
		"ZEBRA=last",
	}

	if diff := cmp.Diff(expected, lines); diff != "" {
		t.Errorf("Format() keys not sorted (-want +got):\n%s", diff)
	}
}

// Example demonstrates basic dotenv formatting
func ExampleDotenvFormatter_Format() {
	formatter := &format.DotenvFormatter{}
	vars := map[string]string{
		"DATABASE_URL": "postgres://user:pass@localhost/db",
		"API_KEY":      "secret-key-123",
		"DEBUG":        "true",
	}

	result, _ := formatter.Format(vars)
	// Note: Output is sorted by key
	// Output will contain:
	// API_KEY=secret-key-123
	// DATABASE_URL=postgres://user:pass@localhost/db
	// DEBUG=true
	_ = result
}

// Example demonstrates escaping of special characters
func ExampleDotenvFormatter_Format_escaping() {
	formatter := &format.DotenvFormatter{}
	vars := map[string]string{
		"QUOTED":    `say "hello"`,
		"SPACED":    "hello world",
		"DOLLAR":    "$HOME/path",
		"BACKSLASH": `path\to\file`,
	}

	result, _ := formatter.Format(vars)
	// Output will contain properly escaped values:
	// BACKSLASH="path\\to\\file"
	// DOLLAR="\$HOME/path"
	// QUOTED="say \"hello\""
	// SPACED="hello world"
	_ = result
}
func TestDotenvFormatter_Format_ErrorHandling(t *testing.T) {
	formatter := &format.DotenvFormatter{}

	tests := []struct {
		name    string
		vars    map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name: "invalid key starting with number",
			vars: map[string]string{
				"123INVALID": "value",
			},
			wantErr: true,
			errMsg:  "invalid key '123INVALID' for dotenv format",
		},
		{
			name: "invalid key with dash",
			vars: map[string]string{
				"KEY-WITH-DASH": "value",
			},
			wantErr: true,
			errMsg:  "invalid key 'KEY-WITH-DASH' for dotenv format",
		},
		{
			name: "invalid key with space",
			vars: map[string]string{
				"KEY WITH SPACE": "value",
			},
			wantErr: true,
			errMsg:  "invalid key 'KEY WITH SPACE' for dotenv format",
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
			name: "valid keys should not error",
			vars: map[string]string{
				"VALID_KEY":    "value1",
				"_PRIVATE_KEY": "value2",
				"KEY123":       "value3",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := formatter.Format(tt.vars)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DotenvFormatter.Format() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("DotenvFormatter.Format() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("DotenvFormatter.Format() unexpected error = %v", err)
			}
		})
	}
}

func TestValidateDotenvKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid key with letters",
			key:     "VALID_KEY",
			wantErr: false,
		},
		{
			name:    "valid key starting with underscore",
			key:     "_PRIVATE_KEY",
			wantErr: false,
		},
		{
			name:    "valid key with numbers",
			key:     "KEY_123",
			wantErr: false,
		},
		{
			name:    "valid lowercase key",
			key:     "lowercase_key",
			wantErr: false,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
			errMsg:  "key cannot be empty",
		},
		{
			name:    "key starting with number",
			key:     "123KEY",
			wantErr: true,
			errMsg:  "key must start with a letter or underscore",
		},
		{
			name:    "key with dash",
			key:     "KEY-WITH-DASH",
			wantErr: true,
			errMsg:  "key contains invalid character '-'",
		},
		{
			name:    "key with space",
			key:     "KEY WITH SPACE",
			wantErr: true,
			errMsg:  "key contains invalid character ' '",
		},
		{
			name:    "key with dot",
			key:     "KEY.WITH.DOT",
			wantErr: true,
			errMsg:  "key contains invalid character '.'",
		},
		{
			name:    "key with special characters",
			key:     "KEY@SYMBOL",
			wantErr: true,
			errMsg:  "key contains invalid character '@'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := format.ValidateDotenvKey(tt.key)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateDotenvKey() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateDotenvKey() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateDotenvKey() unexpected error = %v", err)
			}
		})
	}
}
