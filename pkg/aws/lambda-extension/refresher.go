package extension

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/busser/murmur/pkg/murmur"
)

// Refresher handles periodic secret refresh operations
type Refresher struct {
	config  *ExtensionConfig
	ticker  *time.Ticker
	stopCh  chan struct{}
	errorCh chan error // Send critical errors here when FAIL_ON_REFRESH_ERROR=true
	stopped bool
	started bool
	wg      sync.WaitGroup
}

// NewRefresher creates a new Refresher instance
func NewRefresher(config *ExtensionConfig) *Refresher {
	return &Refresher{
		config:  config,
		stopCh:  make(chan struct{}),
		errorCh: make(chan error, 1), // Buffered to prevent blocking
	}
}

// ErrorCh returns the error channel for critical refresh failures
func (r *Refresher) ErrorCh() <-chan error {
	return r.errorCh
}

// Start begins the periodic refresh ticker
// If refresh interval is 0s, refresh functionality is disabled and this method returns immediately
func (r *Refresher) Start(ctx context.Context) error {
	// Prevent multiple calls to Start()
	if r.started {
		slog.Debug("Refresher already started")
		return nil
	}
	r.started = true

	// Skip ticker creation if refresh is disabled
	if r.config.IsRefreshDisabled() {
		slog.Info("Refresh functionality disabled, skipping ticker creation", "interval", r.config.RefreshInterval.String())
		return nil
	}

	slog.Info("Starting refresh ticker", "interval", r.config.RefreshInterval.String())

	// Create ticker for periodic refresh
	r.ticker = time.NewTicker(r.config.RefreshInterval)

	// Start background goroutine for periodic refresh
	r.wg.Add(1)
	go r.runPeriodicRefresh(ctx)

	return nil
}

// Stop gracefully stops the refresh ticker
func (r *Refresher) Stop() {
	// Prevent multiple calls to Stop()
	if r.stopped {
		slog.Debug("Refresher already stopped")
		return
	}
	r.stopped = true

	slog.Info("Stopping refresh ticker")

	// Signal stop to background goroutine (only if channel is still open)
	select {
	case <-r.stopCh:
		// Channel already closed
	default:
		close(r.stopCh)
	}

	// Stop ticker if it exists
	if r.ticker != nil {
		r.ticker.Stop() // This closes the ticker channel
		slog.Info("Refresh ticker stopped")

		// Wait for background goroutine to finish BEFORE setting ticker to nil
		r.wg.Wait()
		slog.Info("Background goroutine finished")

		// Now it's safe to set ticker to nil since goroutine is done
		r.ticker = nil
	} else {
		slog.Info("No ticker to stop (refresh was disabled)")
	}
}

// checkTTLAndRefresh checks if the secrets file needs refreshing based on TTL and refreshes if needed
func (r *Refresher) checkTTLAndRefresh() error {
	shouldRefresh, err := r.shouldRefresh()
	if err != nil {
		return NewRefreshError("TTL check", err)
	}

	if shouldRefresh {
		slog.Info("Secrets file TTL expired, performing synchronous refresh")
		if err := r.refreshSecrets(); err != nil {
			return NewRefreshError("synchronous refresh", err)
		}
		slog.Info("Synchronous refresh completed successfully")
	}

	return nil
}

// shouldRefresh checks if the secrets file age exceeds the configured TTL
func (r *Refresher) shouldRefresh() (bool, error) {
	// Get file info to check age
	fileInfo, err := os.Stat(r.config.File)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("Secrets file does not exist, refresh needed", "file", r.config.File)
			return true, nil
		}
		return false, NewExtensionError(ErrorTypeRefresh, "failed to stat secrets file", err)
	}

	// Calculate file age
	fileAge := time.Since(fileInfo.ModTime())

	// Check if file age exceeds TTL
	if fileAge > r.config.SecretsTTL {
		slog.Info("Secrets file TTL exceeded", "age", fileAge.String(), "ttl", r.config.SecretsTTL.String())
		return true, nil
	}

	slog.Debug("Secrets file is still fresh", "age", fileAge.String(), "ttl", r.config.SecretsTTL.String())
	return false, nil
}

// runPeriodicRefresh runs the background ticker goroutine for periodic refresh
func (r *Refresher) runPeriodicRefresh(ctx context.Context) {
	defer r.wg.Done()
	slog.Info("Background refresh goroutine started")
	defer slog.Info("Background refresh goroutine stopped")

	// Safety check - ticker should exist when this goroutine is started
	if r.ticker == nil {
		slog.Error("Ticker is nil, exiting background refresh goroutine")
		return
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled, stopping background refresh")
			return
		case <-r.stopCh:
			slog.Info("Stop signal received, stopping background refresh")
			return
		case tick, ok := <-r.ticker.C:
			// Check if ticker channel was closed
			if !ok {
				slog.Info("Ticker channel closed, stopping background refresh")
				return
			}
			_ = tick // Avoid unused variable
			slog.Info("Periodic refresh check triggered")
			if err := r.checkTTLAndRefresh(); err != nil {
				if r.config.FailOnRefreshError {
					slog.Error("Background refresh failed and FAIL_ON_REFRESH_ERROR is true", "error", err)
					// Send critical error to extension via error channel
					select {
					case r.errorCh <- NewRefreshError("background refresh", err):
						// Error sent successfully
					default:
						// Channel full, log and continue
						slog.Error("Error channel full, unable to report critical refresh failure", "error", err)
					}
					return // Exit goroutine on critical failure
				} else {
					slog.Warn("Background refresh failed but continuing operation (FAIL_ON_REFRESH_ERROR=false)", "error", err)
				}
			}
		}
	}
}

// refreshSecrets performs atomic secret refresh using murmur.Export
func (r *Refresher) refreshSecrets() error {
	slog.Info("Starting secret refresh", "file", r.config.File)

	// Use the existing murmur.Export function which handles atomic writes internally
	if err := murmur.Export(r.config.ExportConfig); err != nil {
		return NewExtensionError(ErrorTypeSecretExport, "failed to refresh secrets", err)
	}

	slog.Info("Secret refresh completed successfully", "file", r.config.File)
	return nil
}
