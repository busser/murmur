//go:build e2e

package gcpsm_test

import (
	"testing"
	"time"

	"github.com/busser/whisper/internal/clients/gcpsm"
	"golang.org/x/net/context"
)

func TestClient(t *testing.T) {

	// The secrets this test reads were created with Terraform. The code is in
	// the terraform/layers/gcp-secret-manager directory of this repository.

	client, err := gcpsm.New()
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}

	tt := []struct {
		ref     string
		wantVal string
		wantErr bool
	}{
		{
			ref:     "whisper-tests/secret-sauce",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "whisper-tests/secret-sauce#4",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "whisper-tests/secret-sauce#3",
			wantVal: "ketchup",
			wantErr: false,
		},
		{
			ref:     "whisper-tests/does-not-exist",
			wantVal: "",
			wantErr: true,
		},
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
