package extension

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/busser/murmur/pkg/format"
	"github.com/busser/murmur/pkg/murmur"
)

func TestRefresher_StartStop(t *testing.T) {
	tests := []struct {
		name            string
		refreshInterval time.Duration
		expectTicker    bool
	}{
		{
			name:            "refresh enabled",
			refreshInterval: 1 * time.Minute,
			expectTicker:    true,
		},
		{
			name:            "refresh disabled (zero interval)",
			refreshInterval: 0,
			expectTicker:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ExtensionConfig{
				ExportConfig: murmur.ExportConfig{
					File:      "/tmp/test-secrets.env",
					Formatter: format.Formatters["dotenv"],
					Chmod:     0600,
					Chown:     -1,
				},
				RefreshInterval:    tt.refreshInterval,
				SecretsTTL:         10 * time.Minute,
				FailOnRefreshError: true,
			}

			refresher := NewRefresher(config)

			// Test Start
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := refresher.Start(ctx)
			if err != nil {
				t.Fatalf("Start() failed: %v", err)
			}

			// Check if ticker was created as expected
			if tt.expectTicker && refresher.ticker == nil {
				t.Error("Expected ticker to be created but it was nil")
			}
			if !tt.expectTicker && refresher.ticker != nil {
				t.Error("Expected ticker to be nil but it was created")
			}

			// Test Stop
			refresher.Stop()

			// After stop, ticker should be nil
			if refresher.ticker != nil {
				t.Error("Expected ticker to be nil after Stop()")
			}
		})
	}
}

func TestRefresher_StartStopGraceful(t *testing.T) {
	config := &ExtensionConfig{
		ExportConfig: murmur.ExportConfig{
			File:      "/tmp/test-secrets.env",
			Formatter: format.Formatters["dotenv"],
			Chmod:     0600,
			Chown:     -1,
		},
		RefreshInterval:    500 * time.Millisecond, // Longer interval to avoid race conditions
		SecretsTTL:         10 * time.Minute,
		FailOnRefreshError: false, // Don't fail on errors during test
	}

	refresher := NewRefresher(config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start refresher
	err := refresher.Start(ctx)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Verify ticker was created
	if refresher.ticker == nil {
		t.Error("Expected ticker to be created")
	}

	// Let it run briefly
	time.Sleep(100 * time.Millisecond)

	// Stop should be graceful
	refresher.Stop()

	// Verify ticker is cleaned up
	if refresher.ticker != nil {
		t.Error("Ticker should be nil after Stop()")
	}
}

func TestRefresher_ContextCancellation(t *testing.T) {
	config := &ExtensionConfig{
		ExportConfig: murmur.ExportConfig{
			File:      "/tmp/test-secrets.env",
			Formatter: format.Formatters["dotenv"],
			Chmod:     0600,
			Chown:     -1,
		},
		RefreshInterval:    100 * time.Millisecond, // Short interval for testing
		SecretsTTL:         10 * time.Minute,
		FailOnRefreshError: false,
	}

	refresher := NewRefresher(config)

	ctx, cancel := context.WithCancel(context.Background())

	// Start refresher
	err := refresher.Start(ctx)
	if err != nil {
		t.Fatalf("Start() failed: %v", err)
	}

	// Cancel context to test graceful shutdown
	cancel()

	// Give some time for goroutine to exit
	time.Sleep(50 * time.Millisecond)

	// Stop should still work
	refresher.Stop()
}
func TestRefresher_ShouldRefresh(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "secrets.env")

	tests := []struct {
		name           string
		secretsTTL     time.Duration
		fileAge        time.Duration
		fileExists     bool
		expectedResult bool
		expectError    bool
	}{
		{
			name:           "file does not exist",
			secretsTTL:     10 * time.Minute,
			fileExists:     false,
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "file is fresh (within TTL)",
			secretsTTL:     10 * time.Minute,
			fileAge:        5 * time.Minute,
			fileExists:     true,
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "file is stale (exceeds TTL)",
			secretsTTL:     10 * time.Minute,
			fileAge:        15 * time.Minute,
			fileExists:     true,
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "file age equals TTL",
			secretsTTL:     10 * time.Minute,
			fileAge:        10 * time.Minute,
			fileExists:     true,
			expectedResult: true, // When age >= TTL, should refresh
			expectError:    false,
		},
		{
			name:           "very short TTL with fresh file",
			secretsTTL:     1 * time.Second,
			fileAge:        500 * time.Millisecond,
			fileExists:     true,
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "very short TTL with stale file",
			secretsTTL:     1 * time.Second,
			fileAge:        2 * time.Second,
			fileExists:     true,
			expectedResult: true,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ExtensionConfig{
				ExportConfig: murmur.ExportConfig{
					File:      testFile,
					Formatter: format.Formatters["dotenv"],
					Chmod:     0600,
					Chown:     -1,
				},
				RefreshInterval:    1 * time.Minute,
				SecretsTTL:         tt.secretsTTL,
				FailOnRefreshError: true,
			}

			refresher := NewRefresher(config)

			// Clean up any existing file
			os.Remove(testFile)

			// Create file with specific age if needed
			if tt.fileExists {
				// Create the file
				file, err := os.Create(testFile)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				file.WriteString("TEST_SECRET=test_value\n")
				file.Close()

				// Set the file modification time to simulate age
				pastTime := time.Now().Add(-tt.fileAge)
				err = os.Chtimes(testFile, pastTime, pastTime)
				if err != nil {
					t.Fatalf("Failed to set file time: %v", err)
				}
			}

			// Test shouldRefresh
			result, err := refresher.shouldRefresh()

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check result
			if result != tt.expectedResult {
				t.Errorf("Expected shouldRefresh() = %v, got %v", tt.expectedResult, result)
			}

			// Clean up
			os.Remove(testFile)
		})
	}
}

func TestRefresher_CheckTTLAndRefresh(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "secrets.env")

	tests := []struct {
		name               string
		secretsTTL         time.Duration
		fileAge            time.Duration
		fileExists         bool
		failOnRefreshError bool
		expectRefresh      bool
		expectError        bool
	}{
		{
			name:               "file missing - should refresh",
			secretsTTL:         10 * time.Minute,
			fileExists:         false,
			failOnRefreshError: true,
			expectRefresh:      true,
			expectError:        false, // murmur.Export will handle missing file
		},
		{
			name:               "file fresh - no refresh needed",
			secretsTTL:         10 * time.Minute,
			fileAge:            5 * time.Minute,
			fileExists:         true,
			failOnRefreshError: true,
			expectRefresh:      false,
			expectError:        false,
		},
		{
			name:               "file stale - should refresh",
			secretsTTL:         10 * time.Minute,
			fileAge:            15 * time.Minute,
			fileExists:         true,
			failOnRefreshError: true,
			expectRefresh:      true,
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ExtensionConfig{
				ExportConfig: murmur.ExportConfig{
					File:      testFile,
					Formatter: format.Formatters["dotenv"],
					Chmod:     0600,
					Chown:     -1,
				},
				RefreshInterval:    1 * time.Minute,
				SecretsTTL:         tt.secretsTTL,
				FailOnRefreshError: tt.failOnRefreshError,
			}

			refresher := NewRefresher(config)

			// Clean up any existing file
			os.Remove(testFile)

			// Record initial file state
			var initialModTime time.Time
			if tt.fileExists {
				// Create the file
				file, err := os.Create(testFile)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				file.WriteString("TEST_SECRET=initial_value\n")
				file.Close()

				// Set the file modification time to simulate age
				pastTime := time.Now().Add(-tt.fileAge)
				err = os.Chtimes(testFile, pastTime, pastTime)
				if err != nil {
					t.Fatalf("Failed to set file time: %v", err)
				}

				// Record the initial mod time
				if stat, err := os.Stat(testFile); err == nil {
					initialModTime = stat.ModTime()
				}
			}

			// Test checkTTLAndRefresh
			err := refresher.checkTTLAndRefresh()

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check if refresh occurred by examining file modification time
			if tt.expectRefresh {
				// File should exist after refresh
				stat, err := os.Stat(testFile)
				if err != nil {
					t.Errorf("Expected file to exist after refresh, but got error: %v", err)
				} else {
					// If file existed before, check that mod time changed
					if tt.fileExists && !stat.ModTime().After(initialModTime) {
						t.Error("Expected file modification time to change after refresh")
					}
				}
			} else {
				// If file existed and no refresh was expected, mod time should be unchanged
				if tt.fileExists {
					stat, err := os.Stat(testFile)
					if err != nil {
						t.Errorf("File should still exist: %v", err)
					} else if !stat.ModTime().Equal(initialModTime) {
						t.Error("File modification time should not change when no refresh is needed")
					}
				}
			}

			// Clean up
			os.Remove(testFile)
		})
	}
}

func TestRefresher_FileStatErrors(t *testing.T) {
	// Test handling of file stat errors (e.g., permission denied)
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "secrets.env")

	config := &ExtensionConfig{
		ExportConfig: murmur.ExportConfig{
			File:      testFile,
			Formatter: format.Formatters["dotenv"],
			Chmod:     0600,
			Chown:     -1,
		},
		RefreshInterval:    1 * time.Minute,
		SecretsTTL:         10 * time.Minute,
		FailOnRefreshError: true,
	}

	refresher := NewRefresher(config)

	// Test with non-existent directory (should trigger stat error)
	nonExistentFile := "/non/existent/directory/secrets.env"
	refresher.config.File = nonExistentFile

	// shouldRefresh should handle missing file gracefully
	result, err := refresher.shouldRefresh()
	if err != nil {
		t.Errorf("shouldRefresh should handle missing file gracefully, got error: %v", err)
	}
	if !result {
		t.Error("shouldRefresh should return true for missing file")
	}
}
func TestRefresher_RefreshSecrets(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "secrets.env")

	tests := []struct {
		name            string
		setupExisting   bool
		existingContent string
		expectSuccess   bool
	}{
		{
			name:          "successful refresh to new file",
			setupExisting: false,
			expectSuccess: true,
		},
		{
			name:            "successful refresh replacing existing file",
			setupExisting:   true,
			existingContent: "OLD_SECRET=old_value\n",
			expectSuccess:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ExtensionConfig{
				ExportConfig: murmur.ExportConfig{
					File:      testFile,
					Formatter: format.Formatters["dotenv"],
					Chmod:     0600,
					Chown:     -1,
				},
				RefreshInterval:    1 * time.Minute,
				SecretsTTL:         10 * time.Minute,
				FailOnRefreshError: true,
			}

			refresher := NewRefresher(config)

			// Clean up any existing files
			os.Remove(testFile)

			// Setup existing file if needed
			if tt.setupExisting {
				err := os.WriteFile(testFile, []byte(tt.existingContent), 0600)
				if err != nil {
					t.Fatalf("Failed to create existing file: %v", err)
				}
			}

			// Record initial state
			var initialContent []byte
			var initialModTime time.Time
			if tt.setupExisting {
				initialContent, _ = os.ReadFile(testFile)
				if stat, err := os.Stat(testFile); err == nil {
					initialModTime = stat.ModTime()
				}
			}

			// Test refresh secrets
			err := refresher.refreshSecrets()

			if tt.expectSuccess {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
				}

				// Verify target file exists
				if _, err := os.Stat(testFile); err != nil {
					t.Errorf("Target file should exist after successful refresh: %v", err)
				}

				// Verify file has correct permissions
				if stat, err := os.Stat(testFile); err == nil {
					if stat.Mode().Perm() != 0600 {
						t.Errorf("Expected file permissions 0600, got %o", stat.Mode().Perm())
					}
				}

				// If replacing existing file, verify content changed
				if tt.setupExisting {
					newContent, err := os.ReadFile(testFile)
					if err != nil {
						t.Errorf("Failed to read new file content: %v", err)
					}
					if string(newContent) == string(initialContent) {
						t.Error("File content should have changed after refresh")
					}

					// Verify modification time changed
					if stat, err := os.Stat(testFile); err == nil {
						if !stat.ModTime().After(initialModTime) {
							t.Error("File modification time should be updated after refresh")
						}
					}
				}

				// Verify no temporary files are left behind (atomic behavior)
				tempFile := filepath.Join(filepath.Dir(testFile), "."+filepath.Base(testFile)+".tmp")
				if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
					t.Error("Temporary file should be cleaned up after operation")
				}
			} else {
				if err == nil {
					t.Error("Expected error but got success")
				}
			}

			// Clean up
			os.Remove(testFile)
		})
	}
}

func TestRefresher_RefreshSecretsFailure(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	// Use a non-existent directory to force export failure
	nonExistentDir := filepath.Join(tempDir, "nonexistent")
	testFile := filepath.Join(nonExistentDir, "secrets.env")

	config := &ExtensionConfig{
		ExportConfig: murmur.ExportConfig{
			File:      testFile,
			Formatter: format.Formatters["dotenv"],
			Chmod:     0600,
			Chown:     -1,
		},
		RefreshInterval:    1 * time.Minute,
		SecretsTTL:         10 * time.Minute,
		FailOnRefreshError: true,
	}

	refresher := NewRefresher(config)

	// Test refresh with failure
	err := refresher.refreshSecrets()

	// Should fail due to non-existent directory
	if err == nil {
		t.Error("Expected error due to non-existent directory")
	}

	// Verify no temporary files are left behind (atomic behavior handles cleanup)
	tempFile := filepath.Join(filepath.Dir(testFile), "."+filepath.Base(testFile)+".tmp")
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Error("Temporary file should be cleaned up after failure")
	}
}
