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
	refParts := strings.SplitN(ref, "#", 2)
	if len(refParts) < 1 {
		return "", "", "", errors.New("invalid syntax")
	}
	fullname := refParts[0]
	version = ""
	if len(refParts) == 2 {
		version = refParts[1]
	}

	fullnameParts := strings.SplitN(fullname, "/", 2)
	if len(fullnameParts) < 2 {
		return "", "", "", errors.New("invalid syntax")
	}
	vaultURL = fullnameParts[0]
	name = fullnameParts[1]

	return vaultURL, name, version, nil
}
