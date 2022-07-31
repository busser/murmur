package azkv

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
)

type client struct {
	vaultURI string
	client   *azsecrets.Client
}

// New returns a client that fetches secrets from Azure Key Vault. The vault to
// fetch from must be specified with the WHISPER_AZURE_KEY_VAULT_URL environment
// variable.
func New() (*client, error) {
	vaultURL := os.Getenv("WHISPER_AZURE_KEY_VAULT_URL")
	if vaultURL == "" {
		return nil, errors.New("WHISPER_AZURE_KEY_VAULT_URL must be set")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a credential: %w", err)
	}

	azClient := azsecrets.NewClient(vaultURL, cred, nil)

	return &client{
		client: azClient,
	}, nil
}

func (c *client) Resolve(ctx context.Context, ref string) (string, error) {
	name, version := parseRef(ref)

	// An empty string version gets the latest version of the secret.
	resp, err := c.client.GetSecret(ctx, name, version, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get secret %q version %q: %w", name, version, err)
	}

	return *resp.Value, nil
}

func parseRef(ref string) (name, version string) {
	parts := strings.SplitN(ref, "#", 2)
	if len(parts) < 2 {
		return ref, ""
	}
	return parts[0], parts[1]
}
