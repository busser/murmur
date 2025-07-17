package cmd

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/busser/murmur/pkg/format"
	"github.com/busser/murmur/pkg/murmur"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
)

// Internal struct to capture inputs as-is
type flags struct {
	file   string
	format string
	chmod  string
	chown  string
}

func exportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export resolved secrets to a file",
		Long: `Export resolved secrets to a file instead of passing them through environment variables.
This reduces the risk of accidental secret exposure through process inspection.`,

		Example: `  # Export secrets to default location (/dev/shm/secrets.env):
  murmur export

  # Export to specific file with custom permissions:
  murmur export --file /tmp/secrets.env --chmod 0640

  # Export in Java properties format:
  murmur export --format properties

  # Export with custom ownership (requires root):
  murmur export --chown 1000`,
	}

	flags := flags{}

	// Define flags with environment variable fallbacks
	cmd.Flags().StringVar(&flags.file, "file",
		getEnvWithDefault("MURMUR_EXPORT_FILE", "/dev/shm/secrets.env"),
		"Output file path (env: MURMUR_EXPORT_FILE)")

	cmd.Flags().StringVar(&flags.format, "format",
		getEnvWithDefault("MURMUR_EXPORT_FORMAT", "dotenv"),
		"Output format: dotenv, properties (env: MURMUR_EXPORT_FORMAT)")

	cmd.Flags().StringVar(&flags.chmod, "chmod",
		getEnvWithDefault("MURMUR_EXPORT_CHMOD", "0600"),
		"File permissions in octal format (env: MURMUR_EXPORT_CHMOD)")

	cmd.Flags().StringVar(&flags.chown, "chown",
		getEnvWithDefault("MURMUR_EXPORT_CHOWN", ""),
		"File owner UID (env: MURMUR_EXPORT_CHOWN)")

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		var config *murmur.ExportConfig
		if config, err = newExportConfigFromFlags(&flags); err != nil {
			return
		}

		if err = murmur.Export(*config); err != nil {
			return fmt.Errorf("failed to export secrets: %w", err)
		}

		return
	}

	return cmd
}

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(envVar, defaultValue string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return defaultValue
}

// newExportConfigFromFlags creates and validates an ExportConfig from command flags
func newExportConfigFromFlags(flags *flags) (config *murmur.ExportConfig, err error) {
	// Validate file path
	if flags.file == "" {
		err = multierror.Append(err, fmt.Errorf("file path cannot be empty"))
	}

	// Validate format and get formatter
	formatter, exists := format.Formatters[flags.format]
	if !exists {
		err = multierror.Append(err,
			fmt.Errorf("unsupported format '%s'. Supported formats: %s",
				flags.format,
				strings.Join(slices.Collect(maps.Keys(format.Formatters)), ", ")))
	}

	// Parse/validate chmod
	chmod, parseErr := parseFileMode(flags.chmod)
	if parseErr != nil {
		err = multierror.Append(err, fmt.Errorf("invalid chmod value '%s': %w", flags.chmod, parseErr))
	}

	// Parse/validate chown
	chown, parseErr := parseUID(flags.chown)
	if parseErr != nil {
		err = multierror.Append(err, fmt.Errorf("invalid chown value '%s': %w", flags.chown, parseErr))
	}

	if err == nil {
		config = &murmur.ExportConfig{
			File:      flags.file,
			Formatter: formatter,
			Chmod:     chmod,
			Chown:     chown,
		}
	}

	return
}

// parseFileMode parses octal file permission string
func parseFileMode(modeStr string) (os.FileMode, error) {
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

// parseUID parses user ID string
func parseUID(uidStr string) (int, error) {
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
