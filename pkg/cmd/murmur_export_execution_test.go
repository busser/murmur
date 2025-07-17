package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Helper function to restore environment variables
func restoreEnvironment(originalEnv []string) {
	os.Clearenv()
	for _, env := range originalEnv {
		if idx := strings.Index(env, "="); idx > 0 {
			key := env[:idx]
			value := env[idx+1:]
			os.Setenv(key, value)
		}
	}
}

// TestExportCommandIntegration tests the full CLI command execution flow
func TestExportCommandIntegration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "murmur-export-integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name    string
		args    []string
		envVars map[string]string
		wantErr bool
	}{
		{
			name: "CLI integration with secret filtering",
			args: []string{"export", "--file", filepath.Join(tempDir, "cli.env")},
			envVars: map[string]string{
				"SECRET_VAR": "passthrough:secret-value",
				"NORMAL_VAR": "normal-value",
			},
			wantErr: false,
		},
		{
			name: "CLI validation catches invalid format",
			args: []string{"export", "--file", filepath.Join(tempDir, "invalid.txt"), "--format", "yaml"},
			envVars: map[string]string{
				"SECRET_VAR": "passthrough:secret-value",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalEnv := os.Environ()
			defer restoreEnvironment(originalEnv)

			os.Clearenv()
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Create a mock cobra command to test the export command
			rootCmd := &cobra.Command{Use: "murmur"}
			rootCmd.AddCommand(exportCmd())

			rootCmd.SetArgs(tt.args)
			err := rootCmd.Execute()

			if tt.wantErr && err == nil {
				t.Errorf("Expected error for args %v, got nil", tt.args)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error for args %v: %v", tt.args, err)
			}
		})
	}
}
