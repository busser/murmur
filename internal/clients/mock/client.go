package mock

import (
	"context"
	"fmt"
)

type client struct {
	resolvedRefs []string
}

// New returns a client that fetches no secrets, and simply
func New() (*client, error) {
	return new(client), nil
}

func (c *client) Resolve(ctx context.Context, ref string) (string, error) {
	c.resolvedRefs = append(c.resolvedRefs, ref)

	if ref == "FAIL" {
		return "", ErrorFor(ref)
	}

	return ValueFor(ref), nil
}

func (c *client) ResolvedRefs() []string {
	return c.resolvedRefs
}

func ValueFor(ref string) string {
	return fmt.Sprintf("mock value for ref %q", ref)
}

func ErrorFor(ref string) error {
	return fmt.Errorf("ref %q triggered failure", ref)
}
