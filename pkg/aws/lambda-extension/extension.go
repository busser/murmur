package extension

import (
	"context"
	"log/slog"
	"os"

	"github.com/busser/murmur/pkg/murmur"
)

// Extension represents the Lambda extension with all its components
type Extension struct {
	client    *Client
	config    *ExtensionConfig
	refresher *Refresher
	ctx       context.Context
	cancel    context.CancelFunc
}

// Execute is the main entry point for Lambda extension mode
func Execute() {
	if err := Run(); err != nil {
		slog.Error("Extension failed", "error", err)
		os.Exit(1)
	}
}

// Run orchestrates the extension startup and lifecycle
func Run() error {
	// Create context for the extension lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse configuration from environment variables
	config, err := NewExtensionConfigFromEnv()
	if err != nil {
		slog.Error("Failed to parse configuration", "error", err)
		return NewExtensionError(ErrorTypeConfiguration, "failed to parse configuration", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		slog.Error("Configuration validation failed", "error", err)
		return NewExtensionError(ErrorTypeConfiguration, "configuration validation failed", err)
	}

	// Create Lambda Extensions API client
	runtimeAPI := os.Getenv("AWS_LAMBDA_RUNTIME_API")
	if runtimeAPI == "" {
		err := NewExtensionError(ErrorTypeInit, "AWS_LAMBDA_RUNTIME_API environment variable not set", nil)
		slog.Error("AWS_LAMBDA_RUNTIME_API environment variable not set")
		return err
	}
	client := NewClient(runtimeAPI)

	// Create refresher (will be nil if refresh is disabled)
	var refresher *Refresher
	if !config.IsRefreshDisabled() {
		refresher = NewRefresher(config)
	}

	// Create extension instance
	extension := &Extension{
		client:    client,
		config:    config,
		refresher: refresher,
		ctx:       ctx,
		cancel:    cancel,
	}

	// Start the extension
	return extension.start()
}

// start initializes and registers the extension
func (e *Extension) start() error {
	slog.Info("Starting Lambda extension",
		"file", e.config.File,
		"refresh_interval", e.config.RefreshInterval.String(),
		"ttl", e.config.SecretsTTL.String())

	// Register with Lambda Extensions API
	extensionName := "murmur"
	registerResponse, err := e.client.Register(e.ctx, extensionName)
	if err != nil {
		slog.Error("Failed to register extension", "error", err)
		// Report initialization error to Lambda platform
		if _, initErr := e.client.InitError(e.ctx, string(ErrorTypeRegistration)); initErr != nil {
			slog.Error("Failed to report init error", "error", initErr)
		}
		return NewExtensionError(ErrorTypeRegistration, "extension registration failed", err)
	}

	// Log successful registration
	slog.Info("Extension registered successfully",
		"function_name", registerResponse.FunctionName,
		"function_version", registerResponse.FunctionVersion,
		"handler", registerResponse.Handler)

	// Export secrets immediately after successful registration
	if err := e.exportSecrets(); err != nil {
		slog.Error("Failed to export secrets during startup", "error", err)
		// Report initialization error to Lambda platform
		if _, initErr := e.client.InitError(e.ctx, string(ErrorTypeSecretExport)); initErr != nil {
			slog.Error("Failed to report init error", "error", initErr)
		}
		return NewExtensionError(ErrorTypeSecretExport, "initial secret export failed", err)
	}

	slog.Info("Initial secret export completed successfully", "file", e.config.File)

	// Start background refresh if enabled
	if e.refresher != nil {
		if err := e.refresher.Start(e.ctx); err != nil {
			slog.Error("Failed to start background refresh", "error", err)
			return NewExtensionError(ErrorTypeRefresh, "failed to start background refresh", err)
		}
		slog.Info("Background refresh started", "interval", e.config.RefreshInterval.String())
	}

	// Start the main event loop
	return e.eventLoop()
}

// eventLoop continuously polls for Lambda events and handles them appropriately
func (e *Extension) eventLoop() error {
	slog.Info("Starting event loop")

	// Listen for critical refresh errors from background goroutine
	var refreshErrorCh <-chan error
	if e.refresher != nil {
		refreshErrorCh = e.refresher.ErrorCh()
	}

	for {
		select {
		case <-e.ctx.Done():
			slog.Info("Context cancelled, exiting event loop")
			return e.ctx.Err()

		case refreshErr := <-refreshErrorCh:
			// Critical refresh error from background goroutine
			if refreshErr != nil {
				slog.Error("Critical background refresh error", "error", refreshErr)
				if _, exitErr := e.client.ExitError(e.ctx, string(ErrorTypeRefresh)); exitErr != nil {
					slog.Error("Failed to report exit error", "error", exitErr)
				}
				return NewExtensionError(ErrorTypeRefresh, "critical background refresh failure", refreshErr)
			}

		default:
			// Poll for the next event from Lambda Extensions API
			event, err := e.client.NextEvent(e.ctx)
			if err != nil {
				slog.Error("Error getting next event", "error", err)
				// Report exit error to Lambda platform
				if _, exitErr := e.client.ExitError(e.ctx, string(ErrorTypeEventLoop)); exitErr != nil {
					slog.Error("Failed to report exit error", "error", exitErr)
				}
				return NewExtensionError(ErrorTypeEventLoop, "event loop failed", err)
			}

			slog.Info("Received event",
				"event_type", string(event.EventType),
				"request_id", event.RequestID,
				"deadline_ms", event.DeadlineMs)

			// Handle the event based on its type
			switch event.EventType {
			case Invoke:
				if err := e.handleInvokeEvent(event); err != nil {
					slog.Error("Error handling INVOKE event", "error", err, "request_id", event.RequestID)
					// Continue processing unless configured to fail fast
					if e.config.FailOnRefreshError {
						if _, exitErr := e.client.ExitError(e.ctx, string(ErrorTypeInvokeHandler)); exitErr != nil {
							slog.Error("Failed to report exit error", "error", exitErr)
						}
						return NewExtensionError(ErrorTypeInvokeHandler, "INVOKE event handling failed", err)
					}
				}

			case Shutdown:
				slog.Info("Received SHUTDOWN event, initiating graceful shutdown", "request_id", event.RequestID)
				return e.handleShutdownEvent(event)

			default:
				slog.Warn("Unknown event type", "event_type", string(event.EventType), "request_id", event.RequestID)
			}
		}
	}
}

// handleInvokeEvent processes INVOKE events and checks if secrets need refreshing
func (e *Extension) handleInvokeEvent(event *NextEventResponse) error {
	slog.Info("Processing INVOKE event", "request_id", event.RequestID)

	// Check if we need to refresh secrets based on TTL
	if e.refresher != nil {
		if err := e.refresher.checkTTLAndRefresh(); err != nil {
			slog.Error("TTL-based refresh failed", "error", err, "request_id", event.RequestID)
			return NewRefreshError("TTL-based refresh", err)
		}
	}

	slog.Info("INVOKE event processed successfully", "request_id", event.RequestID)
	return nil
}

// handleShutdownEvent processes SHUTDOWN events and performs cleanup
func (e *Extension) handleShutdownEvent(event *NextEventResponse) error {
	slog.Info("Processing SHUTDOWN event", "request_id", event.RequestID)

	// Stop background refresh ticker if it's running
	if e.refresher != nil {
		slog.Info("Stopping background refresh ticker", "request_id", event.RequestID)
		e.refresher.Stop()
	}

	// Cancel the extension context to signal shutdown to all goroutines
	e.cancel()

	// Clean up any temporary files (this is a placeholder for future implementation)
	// TODO: Implement cleanup of temporary files in task 5.3

	slog.Info("Extension shutdown completed successfully", "request_id", event.RequestID)
	return nil
}

// exportSecrets performs atomic secret export using murmur.Export
func (e *Extension) exportSecrets() error {
	// Use the existing murmur.Export function with our configuration
	return murmur.Export(e.config.ExportConfig)
}
