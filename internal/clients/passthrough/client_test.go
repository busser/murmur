package passthrough

import (
	"context"
	"testing"
)

func TestClient(t *testing.T) {
	tt := []struct {
		ref string
	}{
		{"a"},
		{"b"},
		{"c"},
		{"abc"},
	}

	for _, tc := range tt {
		t.Run(tc.ref, func(t *testing.T) {
			c, err := New()
			if err != nil {
				t.Fatalf("New() returned an error: %v", err)
			}
			defer c.Close()

			val, err := c.Resolve(context.Background(), tc.ref)
			if err != nil {
				t.Fatalf("Resolve() returned an error: %v", err)
			}
			if val != tc.ref {
				t.Errorf("Resolve(%#v) == %#v, want %#v", tc.ref, val, tc.ref)
			}
		})
	}
}
