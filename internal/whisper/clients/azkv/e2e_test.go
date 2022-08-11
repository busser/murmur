//go:build e2e

package azkv_test

import (
	"testing"
	"time"

	"github.com/busser/whisper/internal/whisper/clients/azkv"
	"golang.org/x/net/context"
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
			ref:     "whisper-alpha.vault.azure.net/secret-sauce",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "whisper-alpha.vault.azure.net/secret-sauce#0c2fd54cde7e494faad53882524d358f",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "whisper-alpha.vault.azure.net/secret-sauce#73f5e5ff35a44cdab53b7a34c18da367",
			wantVal: "ketchup",
			wantErr: false,
		},
		{
			ref:     "whisper-alpha.vault.azure.net/does-not-exist",
			wantVal: "",
			wantErr: true,
		},

		// References to the "bravo" vault.
		{
			ref:     "whisper-bravo.vault.azure.net/secret-sauce",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "whisper-bravo.vault.azure.net/secret-sauce#b5f5287b95b24491a7ec5bb6a19ff341",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "whisper-bravo.vault.azure.net/secret-sauce#03bb1bf7a5b44bb28508a6de043faf3c",
			wantVal: "ketchup",
			wantErr: false,
		},
		{
			ref:     "whisper-bravo.vault.azure.net/does-not-exist",
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
