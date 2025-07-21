package murmur

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

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

	return atomicWriteFile(config.File, content, config.Chmod, config.Chown)
}

// atomicWriteFile writes content to a file atomically using a temporary file
func atomicWriteFile(targetPath string, content []byte, chmod os.FileMode, chown int) error {
	// Create temporary file in the same directory as target file
	targetDir := filepath.Dir(targetPath)
	tempFile := filepath.Join(targetDir, "."+filepath.Base(targetPath)+".tmp")

	// Create temporary file with restrictive permissions initially (0600)
	file, err := os.OpenFile(tempFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create temporary secrets file '%s': %w", tempFile, err)
	}

	// Ensure file is closed and cleaned up on error
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close temporary secrets file: %w", closeErr)
		}
		// Clean up temporary file if we're returning an error
		if err != nil {
			if removeErr := os.Remove(tempFile); removeErr != nil && !os.IsNotExist(removeErr) {
				// Log cleanup failure but don't override the original error
			}
		}
	}()

	// Write formatted content to temporary file
	if _, err := file.Write(content); err != nil {
		return fmt.Errorf("failed to write secrets to temporary file '%s': %w", tempFile, err)
	}

	// Close file before changing permissions/ownership
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close temporary secrets file: %w", err)
	}

	// Apply final permissions if different from initial
	if chmod != 0600 {
		if err := os.Chmod(tempFile, chmod); err != nil {
			return fmt.Errorf("failed to set permissions on temporary secrets file '%s': %w", tempFile, err)
		}
	}

	// Apply ownership if specified
	if chown != -1 {
		if err := os.Chown(tempFile, chown, -1); err != nil {
			return fmt.Errorf("failed to change ownership of temporary secrets file '%s' to UID %d: %w", tempFile, chown, err)
		}
	}

	// Atomically move temporary file to target location
	if err := os.Rename(tempFile, targetPath); err != nil {
		return fmt.Errorf("failed to atomically move temporary file to target location '%s': %w", targetPath, err)
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

// ParseFileMode parses octal file permission string
func ParseFileMode(modeStr string) (os.FileMode, error) {
	if modeStr == "" {
		return 0600, nil
	}

	mode, err := strconv.ParseUint(modeStr, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("must be a valid octal number (e.g., 0600, 644)")
	}

	// Validate reasonable permission range
	if mode > 0777 {
		return 0, fmt.Errorf("permissions cannot exceed 0777")
	}

	return os.FileMode(mode), nil
}

// ParseUID parses user ID string
func ParseUID(uidStr string) (int, error) {
	if uidStr == "" {
		return -1, nil // -1 indicates no chown specified
	}

	uid, err := strconv.Atoi(uidStr)
	if err != nil {
		return -1, fmt.Errorf("must be a valid integer (user ID)")
	}

	if uid < 0 {
		return -1, fmt.Errorf("user ID cannot be negative")
	}

	return uid, nil
}
