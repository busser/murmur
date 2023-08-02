//go:build e2e

package azkv_test

import (
	"context"
	"testing"
	"time"

	"github.com/busser/murmur/internal/murmur/providers/azkv"
)

func TestClient(t *testing.T) {

	// The secrets this test reads were created with Terraform. The code is in
	// the terraform/layers/azure-keyvault directory of this repository.

	client, err := azkv.New()
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}

	tt := []struct {
		ref     string
		wantVal string
		wantErr bool
	}{
		// References to the "alpha" vault.
		{
			ref:     "murmur-alpha.vault.azure.net/secret-sauce",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "murmur-alpha.vault.azure.net/secret-sauce#788ffd5cd2224f67b98e12f6fc0cd720",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "murmur-alpha.vault.azure.net/secret-sauce#02fc2105c6b34f8385a2ee8531e4900f",
			wantVal: "ketchup",
			wantErr: false,
		},
		{
			ref:     "murmur-alpha.vault.azure.net/does-not-exist",
			wantVal: "",
			wantErr: true,
		},

		// References to the "bravo" vault.
		{
			ref:     "murmur-bravo.vault.azure.net/secret-sauce",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "murmur-bravo.vault.azure.net/secret-sauce#48b0d307869b4cf9a0141a062ecdc648",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "murmur-bravo.vault.azure.net/secret-sauce#e34b3d09f61f4ed1a1812b88834bcb3e",
			wantVal: "ketchup",
			wantErr: false,
		},
		{
			ref:     "murmur-bravo.vault.azure.net/does-not-exist",
			wantVal: "",
			wantErr: true,
		},

		// Other references.
		{
			ref:     "invalid-ref",
			wantVal: "",
			wantErr: true,
		},
	}

	// Test cases are grouped such that they run in parallel and we can perform
	// cleanup once they are done.
	t.Run("group", func(t *testing.T) {

		for _, tc := range tt {
			tc := tc // capture range variable
			t.Run(tc.ref, func(t *testing.T) {
				t.Parallel()

				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				actualVal, err := client.Resolve(ctx, tc.ref)
				if err != nil && !tc.wantErr {
					t.Errorf("Resolve() returned an error: %v", err)
				}
				if err == nil && tc.wantErr {
					t.Error("Resolve() did not return an error")
				}
				if actualVal != tc.wantVal {
					t.Errorf("Resolve() == %#v, want %#v", actualVal, tc.wantVal)
				}
			})
		}

	})

	if err := client.Close(); err != nil {
		t.Fatalf("Close() returned an error: %v", err)
	}
}
