package whisper

import (
	"context"

	"github.com/busser/whisper/internal/whisper/providers/awssm"
	"github.com/busser/whisper/internal/whisper/providers/azkv"
	"github.com/busser/whisper/internal/whisper/providers/gcpsm"
	"github.com/busser/whisper/internal/whisper/providers/passthrough"
	"github.com/busser/whisper/internal/whisper/providers/scwsm"
)

// A Provider fetches values from a secret store.
type Provider interface {
	// Resolve returns the value of the secret with the given ref. Resolve never
	// gets called after Close.
	Resolve(ctx context.Context, ref string) (string, error)

	// Close signals to the provider that it can release any resources it has
	// allocated, like network connections. Close should return once those
	// resources are released.
	Close() error
}

// A ProviderFactory returns a new Provider.
type ProviderFactory func() (Provider, error)

// ProviderFactories contains a ProviderFactory for each prefix known to
// whisper.
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
