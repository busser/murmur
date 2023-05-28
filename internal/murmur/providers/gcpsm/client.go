package gcpsm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
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
	refParts := strings.SplitN(ref, "#", 2)
	if len(refParts) < 1 {
		return "", "", "", errors.New("invalid syntax")
	}
	fullname := refParts[0]
	version = "latest"
	if len(refParts) == 2 {
		version = refParts[1]
	}

	fullnameParts := strings.SplitN(fullname, "/", 2)
	if len(fullnameParts) < 2 {
		return "", "", "", errors.New("invalid syntax")
	}
	project = fullnameParts[0]
	name = fullnameParts[1]

	return project, name, version, nil
}
