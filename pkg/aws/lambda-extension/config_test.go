package extension

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/busser/murmur/pkg/format"
	"github.com/busser/murmur/pkg/murmur"
	"github.com/google/go-cmp/cmp"
)

func TestNewExtensionConfigFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    *ExtensionConfig
		wantErr bool
	}{
		{
			name:    "default configuration",
			envVars: map[string]string{},
			want: &ExtensionConfig{
				ExportConfig: murmur.ExportConfig{
					File:      "/tmp/secrets.env",
					Formatter: format.Formatters["dotenv"],
					Chmod:     0600,
					Chown:     -1,
				},
				RefreshInterval:    time.Minute,
				SecretsTTL:         10 * time.Minute,
				FailOnRefreshError: true,
			},
			wantErr: false,
		},
		{
			name: "custom configuration",
			envVars: map[string]string{
				"MURMUR_EXPORT_FILE":                  "/custom/path/secrets.env",
				"MURMUR_EXPORT_FORMAT":                "properties",
				"MURMUR_EXPORT_CHMOD":                 "0644",
				"MURMUR_EXPORT_CHOWN":                 "1000",
				"MURMUR_EXPORT_REFRESH_INTERVAL":      "30s",
				"MURMUR_EXPORT_SECRETS_TTL":           "5m",
				"MURMUR_EXPORT_FAIL_ON_REFRESH_ERROR": "false",
			},
			want: &ExtensionConfig{
				ExportConfig: murmur.ExportConfig{
					File:      "/custom/path/secrets.env",
					Formatter: format.Formatters["properties"],
					Chmod:     0644,
					Chown:     1000,
				},
				RefreshInterval:    30 * time.Second,
				SecretsTTL:         5 * time.Minute,
				FailOnRefreshError: false,
			},
			wantErr: false,
		},
		{
			name: "zero refresh interval (disabled refresh)",
			envVars: map[string]string{
				"MURMUR_EXPORT_REFRESH_INTERVAL": "0s",
			},
			want: &ExtensionConfig{
				ExportConfig: murmur.ExportConfig{
					File:      "/tmp/secrets.env",
					Formatter: format.Formatters["dotenv"],
					Chmod:     0600,
					Chown:     -1,
				},
				RefreshInterval:    0,
				SecretsTTL:         10 * time.Minute,
				FailOnRefreshError: true,
			},
			wantErr: false,
		},
		{
			name: "invalid format",
			envVars: map[string]string{
				"MURMUR_EXPORT_FORMAT": "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid refresh interval",
			envVars: map[string]string{
				"MURMUR_EXPORT_REFRESH_INTERVAL": "invalid",
			},
			wantErr: true,
		},
		{
			name: "negative refresh interval",
			envVars: map[string]string{
				"MURMUR_EXPORT_REFRESH_INTERVAL": "-1m",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearEnv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			got, err := NewExtensionConfigFromEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewExtensionConfigFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(ExtensionConfig{})); diff != "" {
					t.Errorf("NewExtensionConfigFromEnv() mismatch (-want +got):\n%s", diff)
				}
			}

			// Clean up environment
			clearEnv()
		})
	}
}

func TestExtensionConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *ExtensionConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			config: &ExtensionConfig{
				ExportConfig: murmur.ExportConfig{
					File:      "/tmp/secrets.env",
					Formatter: format.Formatters["dotenv"],
					Chmod:     0600,
					Chown:     -1,
				},
				RefreshInterval:    time.Minute,
				SecretsTTL:         10 * time.Minute,
				FailOnRefreshError: true,
			},
			wantErr: false,
		},
		{
			name: "zero refresh interval (disabled refresh)",
			config: &ExtensionConfig{
				ExportConfig: murmur.ExportConfig{
					File:      "/tmp/secrets.env",
					Formatter: format.Formatters["dotenv"],
					Chmod:     0600,
					Chown:     -1,
				},
				RefreshInterval:    0,
				SecretsTTL:         10 * time.Minute,
				FailOnRefreshError: true,
			},
			wantErr: false,
		},
		{
			name: "empty file path",
			config: &ExtensionConfig{
				ExportConfig: murmur.ExportConfig{
					File:      "",
					Formatter: format.Formatters["dotenv"],
					Chmod:     0600,
					Chown:     -1,
				},
				RefreshInterval:    time.Minute,
				SecretsTTL:         10 * time.Minute,
				FailOnRefreshError: true,
			},
			wantErr: true,
			errMsg:  "file path cannot be empty",
		},
		{
			name: "TTL less than refresh interval",
			config: &ExtensionConfig{
				ExportConfig: murmur.ExportConfig{
					File:      "/tmp/secrets.env",
					Formatter: format.Formatters["dotenv"],
					Chmod:     0600,
					Chown:     -1,
				},
				RefreshInterval:    time.Minute,
				SecretsTTL:         30 * time.Second, // Less than refresh interval
				FailOnRefreshError: true,
			},
			wantErr: true,
			errMsg:  "secrets TTL (30s) should not be less than refresh interval (1m0s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtensionConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ExtensionConfig.Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestExtensionConfig_IsRefreshDisabled(t *testing.T) {
	tests := []struct {
		name            string
		refreshInterval time.Duration
		want            bool
	}{
		{
			name:            "refresh enabled",
			refreshInterval: time.Minute,
			want:            false,
		},
		{
			name:            "refresh disabled",
			refreshInterval: 0,
			want:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ExtensionConfig{
				RefreshInterval: tt.refreshInterval,
			}
			got := config.IsRefreshDisabled()
			if got != tt.want {
				t.Errorf("ExtensionConfig.IsRefreshDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

// clearEnv clears all MURMUR_EXPORT_* environment variables
func clearEnv() {
	envVars := []string{
		"MURMUR_EXPORT_FILE",
		"MURMUR_EXPORT_FORMAT",
		"MURMUR_EXPORT_CHMOD",
		"MURMUR_EXPORT_CHOWN",
		"MURMUR_EXPORT_REFRESH_INTERVAL",
		"MURMUR_EXPORT_SECRETS_TTL",
		"MURMUR_EXPORT_FAIL_ON_REFRESH_ERROR",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}
