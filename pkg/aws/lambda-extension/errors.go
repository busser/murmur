package extension

import "fmt"

// ExtensionError represents different categories of extension errors
type ExtensionError struct {
	Type    ErrorType
	Message string
	Cause   error
}

// ErrorType categorizes different kinds of extension errors
type ErrorType string

const (
	// ErrorTypeInit indicates initialization failures
	ErrorTypeInit ErrorType = "Extension.InitializationFailed"

	// ErrorTypeRegistration indicates registration failures
	ErrorTypeRegistration ErrorType = "Extension.RegistrationFailed"

	// ErrorTypeSecretExport indicates secret export failures
	ErrorTypeSecretExport ErrorType = "Extension.SecretExportFailed"

	// ErrorTypeRefresh indicates secret refresh failures
	ErrorTypeRefresh ErrorType = "Extension.RefreshFailed"

	// ErrorTypeEventLoop indicates event loop failures
	ErrorTypeEventLoop ErrorType = "Extension.EventLoopError"

	// ErrorTypeInvokeHandler indicates INVOKE event handling failures
	ErrorTypeInvokeHandler ErrorType = "Extension.InvokeHandlerError"

	// ErrorTypeConfiguration indicates configuration errors
	ErrorTypeConfiguration ErrorType = "Extension.ConfigurationError"

	// ErrorTypeNetwork indicates network/API communication errors
	ErrorTypeNetwork ErrorType = "Extension.NetworkError"
)

// Error implements the error interface
func (e *ExtensionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause error
func (e *ExtensionError) Unwrap() error {
	return e.Cause
}

// NewExtensionError creates a new ExtensionError
func NewExtensionError(errorType ErrorType, message string, cause error) *ExtensionError {
	return &ExtensionError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
	}
}

// RefreshError represents a critical refresh failure that should stop the extension
type RefreshError struct {
	Operation string
	Cause     error
}

// Error implements the error interface
func (r *RefreshError) Error() string {
	return fmt.Sprintf("critical refresh failure during %s: %v", r.Operation, r.Cause)
}

// Unwrap returns the underlying cause error
func (r *RefreshError) Unwrap() error {
	return r.Cause
}

// NewRefreshError creates a new RefreshError
func NewRefreshError(operation string, cause error) *RefreshError {
	return &RefreshError{
		Operation: operation,
		Cause:     cause,
	}
}
