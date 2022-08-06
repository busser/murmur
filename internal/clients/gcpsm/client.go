package gcpsm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type client struct {
	gcpClient *secretmanager.Client
}

// New returns a client that fetches secrets from Google Secret Manager.
func New() (*client, error) {
	c, err := secretmanager.NewClient(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to setup client: %w", err)
	}

	return &client{
		gcpClient: c,
	}, nil
}

func (c *client) Resolve(ctx context.Context, ref string) (string, error) {
	project, name, version, err := parseRef(ref)
	if err != nil {
		return "", fmt.Errorf("invalid reference: %w", err)
	}

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: "projects/" + project + "/secrets/" + name + "/versions/" + version,
	}
	resp, err := c.gcpClient.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret %q version %q: %v", name, version, err)
	}

	return string(resp.Payload.Data), nil
}

func (c *client) Close() error {
	return c.gcpClient.Close()
}

func parseRef(ref string) (project, name, version string, err error) {
	parts := strings.SplitN(ref, "/", 3)
	switch {
	case len(parts) < 2:
		return "", "", "", errors.New("not enough information")
	case len(parts) < 3:
		return parts[0], parts[1], "latest", nil
	default:
		return parts[0], parts[1], parts[2], nil
	}
}
