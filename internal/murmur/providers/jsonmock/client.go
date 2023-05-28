package jsonmock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

type client struct {
	mu           sync.RWMutex
	resolvedRefs []string
	closed       bool
}

// New returns a client useful for testing, which provides JSON-encoded values.
func New() *client {
	return new(client)
}

func (c *client) Resolve(ctx context.Context, ref string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.resolvedRefs = append(c.resolvedRefs, ref)

	if ref == "FAIL" {
		return "", ErrorFor(ref)
	}

	return ValueFor(ref), nil
}

func (c *client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return errors.New("already closed")
	}

	c.closed = true

	return nil
}

func (c *client) Closed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.closed
}

func (c *client) ResolvedRefs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

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
