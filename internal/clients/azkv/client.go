package azkv

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
)

type client struct {
	credential azcore.TokenCredential

	mu           sync.RWMutex // Protects keyvaultClients
	vaultClients map[string]*azsecrets.Client
}

// New returns a client that fetches secrets from Azure Key Vault.
func New() (*client, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a credential: %w", err)
	}

	return &client{
		credential:   cred,
		vaultClients: make(map[string]*azsecrets.Client),
	}, nil
}

func (c *client) Resolve(ctx context.Context, ref string) (string, error) {
	vault, name, version, err := parseRef(ref)
	if err != nil {
		return "", fmt.Errorf("invalid reference: %w", err)
	}

	c.createClientIfMissing(vault)

	c.mu.RLock()
	defer c.mu.RUnlock()

	// An empty string version gets the latest version of the secret.
	resp, err := c.vaultClients[vault].GetSecret(ctx, name, version, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get secret %q version %q: %w", name, version, err)
	}

	return *resp.Value, nil
}

func (c *client) Close() error {
	// The client does not need to close its underlying Azure clients.
	// ?(busser): are we sure about this? do any connections need to be closed?
	return nil
}

func (c *client) createClientIfMissing(vault string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.vaultClients[vault] != nil {
		return
	}

	vaultURL := fmt.Sprintf("https://%s/", vault)
	c.vaultClients[vault] = azsecrets.NewClient(vaultURL, c.credential, nil)
}

func parseRef(ref string) (vaultURL, name, version string, err error) {
	parts := strings.SplitN(ref, "/", 3)
	switch {
	case len(parts) < 2:
		return "", "", "", errors.New("not enough information")
	case len(parts) < 3:
		return parts[0], parts[1], "", nil
	default:
		return parts[0], parts[1], parts[2], nil
	}
}
