package murmur

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/busser/murmur/pkg/format"
)

func TestExport(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "murmur-export-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name     string
		vars     map[string]string
		config   ExportConfig
		isRoot   bool
		wantErr  bool
		errMsg   string
		validate func(t *testing.T, filePath string)
	}{
		{
			name: "dotenv format with special characters",
			vars: map[string]string{
				"API_KEY":     "secret123",
				"DB_PASSWORD": "pass with spaces",
			},
			config: ExportConfig{
				File:      filepath.Join(tempDir, "dotenv.env"),
				Formatter: &format.DotenvFormatter{},
				Chmod:     0600,
				Chown:     -1,
			},
			isRoot:  false,
			wantErr: false,
			validate: func(t *testing.T, filePath string) {
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				contentStr := string(content)
				if !strings.Contains(contentStr, "API_KEY=secret123") {
					t.Errorf("Expected API_KEY=secret123 in output, got: %s", contentStr)
				}
				if !strings.Contains(contentStr, `DB_PASSWORD="pass with spaces"`) {
					t.Errorf("Expected quoted DB_PASSWORD in output, got: %s", contentStr)
				}

				info, err := os.Stat(filePath)
				if err != nil {
					t.Fatalf("Failed to stat file: %v", err)
				}
				if info.Mode().Perm() != 0600 {
					t.Errorf("Expected file permissions 0600, got %o", info.Mode().Perm())
				}
			},
		},
		{
			name: "properties format",
			vars: map[string]string{
				"api.key":     "secret123",
				"db.password": "password",
			},
			config: ExportConfig{
				File:      filepath.Join(tempDir, "props.properties"),
				Formatter: &format.PropertiesFormatter{},
				Chmod:     0644,
				Chown:     -1,
			},
			isRoot:  false,
			wantErr: false,
			validate: func(t *testing.T, filePath string) {
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}

				contentStr := string(content)
				if !strings.Contains(contentStr, "api.key = secret123") {
					t.Errorf("Expected api.key = secret123 in output, got: %s", contentStr)
				}
				if !strings.Contains(contentStr, "db.password = password") {
					t.Errorf("Expected db.password = password in output, got: %s", contentStr)
				}

				info, err := os.Stat(filePath)
				if err != nil {
					t.Fatalf("Failed to stat file: %v", err)
				}
				if info.Mode().Perm() != 0644 {
					t.Errorf("Expected file permissions 0644, got %o", info.Mode().Perm())
				}
			},
		},
		{
			name: "empty secrets map creates empty file",
			vars: map[string]string{},
			config: ExportConfig{
				File:      filepath.Join(tempDir, "empty.env"),
				Formatter: &format.DotenvFormatter{},
				Chmod:     0600,
				Chown:     -1,
			},
			isRoot:  false,
			wantErr: false,
			validate: func(t *testing.T, filePath string) {
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}
				if len(content) != 0 {
					t.Errorf("Expected empty file, got content: %s", string(content))
				}
			},
		},
		{
			name: "chown requires root privileges",
			vars: map[string]string{"KEY": "value"},
			config: ExportConfig{
				File:      filepath.Join(tempDir, "chown-fail.env"),
				Formatter: &format.DotenvFormatter{},
				Chmod:     0600,
				Chown:     1000,
			},
			isRoot:  false,
			wantErr: true,
			errMsg:  "chown operation requires root privileges",
		},
		{
			name: "chown passes root check but may fail actual operation",
			vars: map[string]string{"KEY": "value"},
			config: ExportConfig{
				File:      filepath.Join(tempDir, "chown-success.env"),
				Formatter: &format.DotenvFormatter{},
				Chmod:     0600,
				Chown:     1000, // Non-root UID to avoid actual root operations
			},
			isRoot:  true,
			wantErr: true, // Expect chown to fail in test environment
			errMsg:  "failed to change ownership",
		},
		{
			name: "invalid directory error",
			vars: map[string]string{"KEY": "value"},
			config: ExportConfig{
				File:      "/nonexistent/directory/file.env",
				Formatter: &format.DotenvFormatter{},
				Chmod:     0600,
				Chown:     -1,
			},
			isRoot:  false,
			wantErr: true,
			errMsg:  "directory validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := export(tt.config, tt.isRoot, tt.vars)

			if tt.wantErr {
				if err == nil {
					t.Errorf("export() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("export() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("export() unexpected error = %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, tt.config.File)
			}
		})
	}
}

// TestPublicExportAPI provides a simple smoke test for the public Export method
// Most testing is done on the private export method for better control
func TestPublicExportAPI(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "murmur-public-export-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up a simple environment variable that doesn't require secret resolution
	os.Setenv("MURMUR_TEST_VAR", "test-value")
	defer os.Unsetenv("MURMUR_TEST_VAR")

	config := ExportConfig{
		File:      filepath.Join(tempDir, "public-api-test.env"),
		Formatter: &format.DotenvFormatter{},
		Chmod:     0600,
		Chown:     -1,
	}

	// This tests the public API that handles environment resolution internally
	err = Export(config)
	if err != nil {
		t.Errorf("Public Export() unexpected error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(config.File); err != nil {
		t.Errorf("Expected file to be created: %v", err)
	}
}

func TestValidateDirectoryExists(t *testing.T) {
	// Create temporary directory for tests
	tempDir, err := os.MkdirTemp("", "murmur-dir-validation-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name     string
		filePath string
		setup    func() error
		cleanup  func() error
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid directory",
			filePath: filepath.Join(tempDir, "valid.env"),
			wantErr:  false,
		},
		{
			name:     "nonexistent directory",
			filePath: "/nonexistent/directory/file.env",
			wantErr:  true,
			errMsg:   "output directory '/nonexistent/directory' does not exist",
		},
		{
			name:     "directory is actually a file",
			filePath: filepath.Join(tempDir, "notdir", "file.env"),
			setup: func() error {
				// Create a file where we expect a directory
				return os.WriteFile(filepath.Join(tempDir, "notdir"), []byte("content"), 0644)
			},
			cleanup: func() error {
				return os.Remove(filepath.Join(tempDir, "notdir"))
			},
			wantErr: true,
			errMsg:  "exists but is not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			if tt.cleanup != nil {
				defer func() {
					if err := tt.cleanup(); err != nil {
						t.Logf("Cleanup failed: %v", err)
					}
				}()
			}

			err := validateDirectoryExists(tt.filePath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateDirectoryExists() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateDirectoryExists() error = %v, want error containing %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("validateDirectoryExists() unexpected error = %v", err)
			}
		})
	}
}

// TestFormatValidation tests format-specific validation errors
func TestFormatValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "murmur-format-validation-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name    string
		vars    map[string]string
		config  ExportConfig
		isRoot  bool
		wantErr bool
		errMsg  string
	}{
		{
			name: "invalid dotenv key format",
			vars: map[string]string{
				"123INVALID": "value", // Key starting with number
			},
			config: ExportConfig{
				File:      filepath.Join(tempDir, "invalid-key.env"),
				Formatter: &format.DotenvFormatter{},
				Chmod:     0600,
				Chown:     -1,
			},
			isRoot:  false,
			wantErr: true,
			errMsg:  "invalid key '123INVALID' for dotenv format",
		},
		{
			name: "invalid properties key with equals",
			vars: map[string]string{
				"KEY=WITH=EQUALS": "value",
			},
			config: ExportConfig{
				File:      filepath.Join(tempDir, "invalid-props-equals.properties"),
				Formatter: &format.PropertiesFormatter{},
				Chmod:     0600,
				Chown:     -1,
			},
			isRoot:  false,
			wantErr: true,
			errMsg:  "cannot contain '=' or ':' characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := export(tt.config, tt.isRoot, tt.vars)

			if tt.wantErr {
				if err == nil {
					t.Errorf("export() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("export() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("export() unexpected error = %v", err)
			}
		})
	}
}
