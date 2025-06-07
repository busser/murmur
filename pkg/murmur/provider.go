package murmur

import (
	"context"

	"github.com/busser/murmur/pkg/murmur/providers/awssm"
	"github.com/busser/murmur/pkg/murmur/providers/azkv"
	"github.com/busser/murmur/pkg/murmur/providers/gcpsm"
	"github.com/busser/murmur/pkg/murmur/providers/passthrough"
	"github.com/busser/murmur/pkg/murmur/providers/scwsm"
)

// Provider fetches values from a secret store (e.g., AWS Secrets Manager, Azure Key Vault).
// Each provider implements a specific cloud secret management service.
//
// Providers are designed to be:
//   - Thread-safe for concurrent Resolve calls
//   - Stateful (may hold connections, credentials)
//   - Resource-managed (must call Close when done)
//
// Example usage:
//
//	provider, err := awssm.New()
//	if err != nil {
//	    return err
//	}
//	defer provider.Close()
//
//	secret, err := provider.Resolve(ctx, "my-secret#AWSCURRENT")
//	if err != nil {
//	    return err
//	}
type Provider interface {
	// Resolve returns the value of the secret with the given ref.
	// The ref format is provider-specific (e.g., "secret-name#version").
	// Resolve should be thread-safe and may be called concurrently.
	// Resolve will never be called after Close.
	Resolve(ctx context.Context, ref string) (string, error)

	// Close signals to the provider that it can release any resources it has
	// allocated, like network connections. Close should return once those
	// resources are released. After Close is called, Resolve should not be called.
	Close() error
}

// ProviderFactory creates a new Provider instance.
// Each call should return a fresh provider with its own resources.
type ProviderFactory func() (Provider, error)

// ProviderFactories contains a ProviderFactory for each provider prefix known to murmur.
// This map is used by the resolution pipeline to create providers on-demand.
//
// Available providers:
//   - "awssm": AWS Secrets Manager
//   - "azkv": Azure Key Vault
//   - "gcpsm": GCP Secret Manager  
//   - "scwsm": Scaleway Secret Manager
//   - "passthrough": Testing/no-op provider
var ProviderFactories = map[string]ProviderFactory{
	// Passthrough
	"passthrough": func() (Provider, error) { return passthrough.New() },
	// Azure Key Vault
	"azkv": func() (Provider, error) { return azkv.New() },
	// Google Cloud Secret Manager
	"gcpsm": func() (Provider, error) { return gcpsm.New() },
	// AWS Secrets Manager
	"awssm": func() (Provider, error) { return awssm.New() },
	// Scaleway Secret Manager
	"scwsm": func() (Provider, error) { return scwsm.New() },
}
