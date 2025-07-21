package extension

import (
	"fmt"
	"strconv"
	"time"

	"github.com/busser/murmur/pkg/environ"
	"github.com/busser/murmur/pkg/format"
	"github.com/busser/murmur/pkg/murmur"
	"github.com/hashicorp/go-multierror"
)

// ExtensionConfig holds configuration for the Lambda extension
type ExtensionConfig struct {
	// Embedded export configuration (reuses murmur.ExportConfig)
	murmur.ExportConfig

	// Extension-specific configuration
	RefreshInterval    time.Duration // How often to check for refresh
	SecretsTTL         time.Duration // How long secrets are considered fresh
	FailOnRefreshError bool          // Whether to exit on refresh failures
}

// NewExtensionConfigFromEnv creates a new Config from environment variables
func NewExtensionConfigFromEnv() (*ExtensionConfig, error) {
	var err error

	// Parse export configuration from environment variables
	exportConfig, parseErr := parseExportConfig()
	if parseErr != nil {
		err = multierror.Append(err, parseErr)
	}

	// Parse extension-specific configuration
	extensionConfig, parseErr := parseExtensionConfig(exportConfig)
	if parseErr != nil {
		err = multierror.Append(err, parseErr)
	}

	if err != nil {
		return nil, fmt.Errorf("configuration parsing failed: %w", err)
	}

	config := &ExtensionConfig{
		ExportConfig:       *exportConfig,
		RefreshInterval:    extensionConfig.RefreshInterval,
		SecretsTTL:         extensionConfig.SecretsTTL,
		FailOnRefreshError: extensionConfig.FailOnRefreshError,
	}

	return config, nil
}

// parseExportConfig parses the export configuration from environment variables
func parseExportConfig() (*murmur.ExportConfig, error) {
	var err error

	// Get file path with Lambda-appropriate default
	file := environ.GetEnvWithDefault("MURMUR_EXPORT_FILE", "/tmp/secrets.env")
	if file == "" {
		err = multierror.Append(err, fmt.Errorf("file path cannot be empty"))
	}

	// Get format with default
	formatStr := environ.GetEnvWithDefault("MURMUR_EXPORT_FORMAT", "dotenv")
	formatter, exists := format.Formatters[formatStr]
	if !exists {
		err = multierror.Append(err, fmt.Errorf("unsupported format '%s'", formatStr))
	}

	// Parse chmod with default
	chmodStr := environ.GetEnvWithDefault("MURMUR_EXPORT_CHMOD", "0600")
	chmod, parseErr := murmur.ParseFileMode(chmodStr)
	if parseErr != nil {
		err = multierror.Append(err, fmt.Errorf("invalid chmod value '%s': %w", chmodStr, parseErr))
	}

	// Parse chown (optional)
	chownStr := environ.GetEnvWithDefault("MURMUR_EXPORT_CHOWN", "")
	chown, parseErr := murmur.ParseUID(chownStr)
	if parseErr != nil {
		err = multierror.Append(err, fmt.Errorf("invalid chown value '%s': %w", chownStr, parseErr))
	}

	if err != nil {
		return nil, err
	}

	return &murmur.ExportConfig{
		File:      file,
		Formatter: formatter,
		Chmod:     chmod,
		Chown:     chown,
	}, nil
}

// extensionConfig holds extension-specific configuration values
type extensionConfig struct {
	RefreshInterval    time.Duration
	SecretsTTL         time.Duration
	FailOnRefreshError bool
}

// parseExtensionConfig parses extension-specific configuration from environment variables
func parseExtensionConfig(exportConfig *murmur.ExportConfig) (*extensionConfig, error) {
	var err error

	// Parse refresh interval
	refreshIntervalStr := environ.GetEnvWithDefault("MURMUR_EXPORT_REFRESH_INTERVAL", "1m")
	refreshInterval, parseErr := parseRefreshInterval(refreshIntervalStr)
	if parseErr != nil {
		err = multierror.Append(err, parseErr)
	}

	// Parse secrets TTL
	secretsTTLStr := environ.GetEnvWithDefault("MURMUR_EXPORT_SECRETS_TTL", "10m")
	secretsTTL, parseErr := parseSecretsTTL(secretsTTLStr)
	if parseErr != nil {
		err = multierror.Append(err, parseErr)
	}

	// Parse fail on refresh error
	failOnRefreshErrorStr := environ.GetEnvWithDefault("MURMUR_EXPORT_FAIL_ON_REFRESH_ERROR", "true")
	failOnRefreshError, parseErr := strconv.ParseBool(failOnRefreshErrorStr)
	if parseErr != nil {
		err = multierror.Append(err, fmt.Errorf("invalid fail on refresh error value '%s': must be true or false", failOnRefreshErrorStr))
	}

	if err != nil {
		return nil, err
	}

	return &extensionConfig{
		RefreshInterval:    refreshInterval,
		SecretsTTL:         secretsTTL,
		FailOnRefreshError: failOnRefreshError,
	}, nil
}

// parseRefreshInterval parses and validates refresh interval duration
// Zero duration (0s) is allowed and disables all refresh functionality
func parseRefreshInterval(intervalStr string) (time.Duration, error) {
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		return 0, fmt.Errorf("invalid refresh interval '%s': %w", intervalStr, err)
	}

	if interval < 0 {
		return 0, fmt.Errorf("refresh interval cannot be negative, got %v", interval)
	}

	return interval, nil
}

// parseSecretsTTL parses and validates secrets TTL duration
func parseSecretsTTL(ttlStr string) (time.Duration, error) {
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		return 0, fmt.Errorf("invalid secrets TTL '%s': %w", ttlStr, err)
	}

	if ttl <= 0 {
		return 0, fmt.Errorf("secrets TTL must be positive, got %v", ttl)
	}

	return ttl, nil
}

// Validate performs comprehensive validation of the configuration
func (c *ExtensionConfig) Validate() error {
	var err error

	// Validate export configuration
	if validateErr := c.validateExportConfig(); validateErr != nil {
		err = multierror.Append(err, validateErr)
	}

	// Validate extension-specific configuration
	if validateErr := c.validateExtensionConfig(); validateErr != nil {
		err = multierror.Append(err, validateErr)
	}

	return err
}

// validateExportConfig validates the embedded export configuration
func (c *ExtensionConfig) validateExportConfig() error {
	var err error

	// Validate file path
	if c.File == "" {
		err = multierror.Append(err, fmt.Errorf("file path cannot be empty"))
	}

	// Validate formatter is set
	if c.Formatter == nil {
		err = multierror.Append(err, fmt.Errorf("formatter cannot be nil"))
	}

	// Validate file permissions are reasonable
	if c.Chmod > 0777 {
		err = multierror.Append(err, fmt.Errorf("file permissions cannot exceed 0777"))
	}

	// Validate chown if specified
	if c.Chown < -1 {
		err = multierror.Append(err, fmt.Errorf("chown UID cannot be less than -1"))
	}

	return err
}

// validateExtensionConfig validates extension-specific configuration
func (c *ExtensionConfig) validateExtensionConfig() error {
	var err error

	// Validate refresh interval (zero is allowed to disable refresh)
	if c.RefreshInterval < 0 {
		err = multierror.Append(err, fmt.Errorf("refresh interval cannot be negative, got %v", c.RefreshInterval))
	}

	// Validate secrets TTL
	if c.SecretsTTL <= 0 {
		err = multierror.Append(err, fmt.Errorf("secrets TTL must be positive, got %v", c.SecretsTTL))
	}

	// Only validate TTL relationship if refresh is enabled (non-zero)
	if c.RefreshInterval > 0 && c.SecretsTTL < c.RefreshInterval {
		err = multierror.Append(err, fmt.Errorf("secrets TTL (%v) should not be less than refresh interval (%v)", c.SecretsTTL, c.RefreshInterval))
	}

	return err
}

// IsRefreshDisabled returns true if refresh functionality is disabled (refresh interval is 0)
func (c *ExtensionConfig) IsRefreshDisabled() bool {
	return c.RefreshInterval == 0
}
