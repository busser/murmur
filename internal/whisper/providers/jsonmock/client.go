package jsonmock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type client struct {
	resolvedRefs []string
	closed       bool
}

// New returns a client useful for testing, which provides JSON-encoded values.
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

func (c *client) Close() error {
	if c.closed {
		return errors.New("already closed")
	}

	c.closed = true

	return nil
}

func (c *client) Closed() bool {
	return c.closed
}

func (c *client) ResolvedRefs() []string {
	return c.resolvedRefs
}

const Key = "ref"

func ValueFor(ref string) string {
	obj := map[string]string{
		Key: ref,
	}

	encoded, err := json.Marshal(obj)
	if err != nil {
		panic(fmt.Sprintf("could not encode with ref %q", ref))
	}

	return string(encoded)
}

func ErrorFor(ref string) error {
	return fmt.Errorf("ref %q triggered failure", ref)
}
