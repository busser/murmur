package passthrough

import (
	"context"
)

type client struct{}

// New returns a client that fetches no secrets, and simply uses the secret's
// reference as its value.
func New() (*client, error) {
	return &client{}, nil
}

func (c *client) Resolve(ctx context.Context, ref string) (string, error) {
	return ref, nil
}
