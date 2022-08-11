//go:build e2e

package awssm_test

import (
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/busser/whisper/internal/whisper/providers/awssm"
)

func TestClient(t *testing.T) {

	// The secrets this test reads were created with Terraform. The code is in
	// the terraform/layers/aws-secrets-manager directory of this repository.

	client, err := awssm.New()
	if err != nil {
		t.Fatalf("New() returned an error: %v", err)
	}

	tt := []struct {
		ref     string
		wantVal string
		wantErr bool
	}{
		// References by name.
		{
			ref:     "secret-sauce",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "secret-sauce#AWSCURRENT",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "secret-sauce#9AF93B18-59D6-4C19-92AC-3F69A115D404",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "secret-sauce#v2",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "secret-sauce#97DD35A4-DD9B-4E4B-B371-9F2CA4673A41",
			wantVal: "ketchup",
			wantErr: false,
		},
		{
			ref:     "secret-sauce#v1",
			wantVal: "ketchup",
			wantErr: false,
		},
		{
			ref:     "does-not-exist",
			wantVal: "",
			wantErr: true,
		},

		// References by ARN.
		{
			ref:     "arn:aws:secretsmanager:eu-west-3:531255069405:secret:secret-sauce-sWcbiZ",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "arn:aws:secretsmanager:eu-west-3:531255069405:secret:secret-sauce-sWcbiZ#AWSCURRENT",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "arn:aws:secretsmanager:eu-west-3:531255069405:secret:secret-sauce-sWcbiZ#9AF93B18-59D6-4C19-92AC-3F69A115D404",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "arn:aws:secretsmanager:eu-west-3:531255069405:secret:secret-sauce-sWcbiZ#v2",
			wantVal: "szechuan",
			wantErr: false,
		},
		{
			ref:     "arn:aws:secretsmanager:eu-west-3:531255069405:secret:secret-sauce-sWcbiZ#97DD35A4-DD9B-4E4B-B371-9F2CA4673A41",
			wantVal: "ketchup",
			wantErr: false,
		},
		{
			ref:     "arn:aws:secretsmanager:eu-west-3:531255069405:secret:secret-sauce-sWcbiZ#v1",
			wantVal: "ketchup",
			wantErr: false,
		},
		{
			ref:     "arn:aws:secretsmanager:eu-west-3:531255069405:secret:does-not-exist",
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
