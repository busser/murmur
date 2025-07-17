package murmur

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/busser/murmur/pkg/environ"
	"github.com/busser/murmur/pkg/format"
)

// ExportConfig holds configuration for secrets export operations
type ExportConfig struct {
	File      string           // Output file path
	Formatter format.Formatter // Output formatter instance
	Chmod     os.FileMode      // File permissions
	Chown     int              // File owner UID (-1 if not specified)
}

// Export writes resolved secrets to a file with specified configuration
// This is the public API that handles environmental concerns
func Export(config ExportConfig) error {
	envVars := environ.ToMap(os.Environ())
	isRoot := os.Geteuid() == 0

	// Resolve secrets using ResolveSecrets (secrets-only resolution)
	resolvedSecrets, err := ResolveSecrets(envVars)
	if err != nil {
		return fmt.Errorf("secret resolution failed: %w", err)
	}

	return export(config, isRoot, resolvedSecrets)
}

// export is the private method containing pure logic, easily testable
func export(config ExportConfig, isRoot bool, vars map[string]string) error {
	// Early validation: check root privileges if chown is requested
	if config.Chown != -1 && !isRoot {
		return fmt.Errorf("chown operation requires root privileges")
	}

	// Validate directory exists before attempting file creation
	if err := validateDirectoryExists(config.File); err != nil {
		return fmt.Errorf("directory validation failed: %w", err)
	}

	// Convert secrets to specified format with enhanced error handling
	content, err := config.Formatter.Format(vars)
	if err != nil {
		return fmt.Errorf("failed to format secrets: %w", err)
	}

	// Create file with restrictive permissions initially (0600)
	file, err := os.OpenFile(config.File, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create secrets file '%s': %w", config.File, err)
	}

	// Ensure file is closed and cleaned up on error
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close secrets file: %w", closeErr)
		}
	}()

	// Write formatted content
	if _, err := file.Write(content); err != nil {
		// Clean up file on write failure
		os.Remove(config.File)
		return fmt.Errorf("failed to write secrets to file '%s': %w", config.File, err)
	}

	// Close file before changing permissions/ownership
	if err := file.Close(); err != nil {
		os.Remove(config.File)
		return fmt.Errorf("failed to close secrets file: %w", err)
	}

	// Apply final permissions if different from initial
	if config.Chmod != 0600 {
		if err := os.Chmod(config.File, config.Chmod); err != nil {
			os.Remove(config.File)
			return fmt.Errorf("failed to set permissions on secrets file '%s': %w", config.File, err)
		}
	}

	// Apply ownership if specified (root check already done early)
	if config.Chown != -1 {
		if err := os.Chown(config.File, config.Chown, -1); err != nil {
			os.Remove(config.File)
			return fmt.Errorf("failed to change ownership of secrets file '%s' to UID %d: %w", config.File, config.Chown, err)
		}
	}

	return nil
}

// validateDirectoryExists checks if the parent directory for the given file path exists
func validateDirectoryExists(filePath string) error {
	parentDir := filepath.Dir(filePath)

	// Check if directory exists
	parentStat, err := os.Stat(parentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("output directory '%s' does not exist", parentDir)
		}
		return fmt.Errorf("failed to check directory '%s': %w", parentDir, err)
	}

	// Check if it's actually a directory
	if !parentStat.IsDir() {
		return fmt.Errorf("output path '%s' exists but is not a directory", parentDir)
	}

	return nil
}
