package whisper

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"

	"github.com/busser/whisper/internal/clients/awssm"
	"github.com/busser/whisper/internal/clients/azkv"
	"github.com/busser/whisper/internal/clients/gcpsm"
	"github.com/busser/whisper/internal/clients/passthrough"
)

// A Client fetches values from a secret store.
type Client interface {
	// Resolve returns the value of the secret with the given ref. Resolve never
	// gets called after Close.
	Resolve(ctx context.Context, ref string) (string, error)

	// Close signals to the client that it can release any resources it has
	// allocated, like network connections. Close should return once those
	// resources are released.
	Close() error
}

// A ClientFactory returns a new Client.
type ClientFactory func() (Client, error)

// ClientFactories contains a ClientFactory for each prefix known to whisper.
var ClientFactories = map[string]ClientFactory{
	// Passthrough
	"passthrough": func() (Client, error) { return passthrough.New() },
	// Azure Key Vault
	"azkv": func() (Client, error) { return azkv.New() },
	// Google Cloud Secret Manager
	"gcpsm": func() (Client, error) { return gcpsm.New() },
	// AWS Secrets Manager
	"awssm": func() (Client, error) { return awssm.New() },
}

// ResolveAll returns a map with the same keys as vars, where all values with
// known prefixes have been replaced with their values.
func ResolveAll(vars map[string]string) (map[string]string, error) {

	// First, group variables based on the prefix of their values. This prefix
	// will serve to instantiate the necessary clients.

	varsByPrefix := make(map[string][]string)
	for k, v := range vars {
		split := strings.SplitN(v, ":", 2)

		// If the variable contains no colons, then it has no prefix and whisper
		// should ignore it.
		if len(split) < 2 {
			continue
		}

		prefix := split[0]

		// If the variable has an unknown prefix, then whisper should ignore it.
		if _, known := ClientFactories[prefix]; !known {
			continue
		}

		varsByPrefix[prefix] = append(varsByPrefix[prefix], k)
	}

	var err error

	clientByPrefix := make(map[string]Client)
	for prefix := range varsByPrefix {
		newClient, known := ClientFactories[prefix]
		if !known {
			err = multierror.Append(err, fmt.Errorf("no client for prefix %q", prefix))
			continue
		}

		client, clientErr := newClient()
		if clientErr != nil {
			err = multierror.Append(err, fmt.Errorf("client for %q: %w", prefix, clientErr))
			continue
		}

		clientByPrefix[prefix] = client

		defer client.Close() // TODO(busser): handle error (log it?)
	}
	if err != nil {
		return nil, err
	}

	newVars := make(map[string]string)
	for prefix, keys := range varsByPrefix {
		c := clientByPrefix[prefix]
		for _, k := range keys {
			ref := strings.TrimPrefix(vars[k], prefix+":")
			val, resolveErr := c.Resolve(context.TODO(), ref)
			if resolveErr != nil {
				err = multierror.Append(err, fmt.Errorf("%s: %w", k, resolveErr))
				continue
			}
			newVars[k] = val
		}
	}
	if err != nil {
		return nil, err
	}

	mergedVars := make(map[string]string)
	for k := range vars {
		if v, ok := newVars[k]; ok {
			mergedVars[k] = v
			continue
		}
		mergedVars[k] = vars[k]
	}

	return mergedVars, nil
}
