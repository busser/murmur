//go:build e2e

package scwsm_test

import (
	"testing"
	"time"

	"github.com/busser/whisper/internal/whisper/providers/scwsm"
	"golang.org/x/net/context"
)

func TestClient(t *testing.T) {

	// The secrets this test reads were created with Terraform. The code is in
	// the terraform/layers/gcp-secret-manager directory of this repository.

	client, err := scwsm.New()
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}

	tt := []struct {
		ref     string
		wantVal string
		wantErr bool
	}{
		{
			ref:     "secret-sauce",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "secret-sauce#2",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "secret-sauce#1",
			wantVal: "ketchup",
			wantErr: false,
		},
		{
			ref:     "fr-par/secret-sauce",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "fr-par/secret-sauce#2",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "fr-par/secret-sauce#1",
			wantVal: "ketchup",
			wantErr: false,
		},
		{
			ref:     "3f34b83f-47a6-4344-bcd4-b63721481cd3",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "3f34b83f-47a6-4344-bcd4-b63721481cd3#2",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "3f34b83f-47a6-4344-bcd4-b63721481cd3#1",
			wantVal: "ketchup",
			wantErr: false,
		},
		{
			ref:     "fr-par/3f34b83f-47a6-4344-bcd4-b63721481cd3",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "fr-par/3f34b83f-47a6-4344-bcd4-b63721481cd3#2",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "fr-par/3f34b83f-47a6-4344-bcd4-b63721481cd3#1",
			wantVal: "ketchup",
			wantErr: false,
		},
		{
			ref:     "does-not-exist",
			wantVal: "",
			wantErr: true,
		},
		{
			ref:     "fr-par/does-not-exist",
			wantVal: "",
			wantErr: true,
		},
		{
			ref:     "fr-par/does-not-exist#123",
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
