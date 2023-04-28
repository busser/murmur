//go:build e2e

package whisper

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestResolveAllEndToEnd(t *testing.T) {

	// The secrets this test reads were created with Terraform. The code is in
	// the terraform directory of this repository.

	// This test does a high-level pass using all clients. More in-depth
	// end-to-end testing of each provider is done in the provider's package.

	envVars := map[string]string{
		"NOT_A_SECRET":        "My app listens on port 3000",
		"FROM_AZURE":          "azkv:whisper-alpha.vault.azure.net/secret-sauce",
		"FROM_AWS":            "awssm:secret-sauce",
		"FROM_GCP":            "gcpsm:whisper-tests/secret-sauce",
		"FROM_SCALEWAY":       "scwsm:secret-sauce",
		"FROM_PASSTHROUGH":    "passthrough:szechuan",
		"JSON_SECRET":         `passthrough:{"sauce": "szechuan"}|jsonpath:{ .sauce }`,
		"LOOKS_LIKE_A_SECRET": "baz:but isn't a secret",
	}

	actual, err := ResolveAll(envVars)
	if err != nil {
		t.Fatalf("ResolveAll() returned an error: %v", err)
	}

	want := map[string]string{
		"NOT_A_SECRET":        "My app listens on port 3000",
		"FROM_AZURE":          "szechuan",
		"FROM_AWS":            "szechuan",
		"FROM_GCP":            "szechuan",
		"FROM_SCALEWAY":       "szechuan",
		"FROM_PASSTHROUGH":    "szechuan",
		"JSON_SECRET":         `szechuan`,
		"LOOKS_LIKE_A_SECRET": "baz:but isn't a secret",
	}

	if diff := cmp.Diff(want, actual); diff != "" {
		t.Errorf("ResolveAll() mismatch (-want +got):\n%s", diff)
	}
}

func TestResolveAllEndToEndWithError(t *testing.T) {
	envVars := map[string]string{
		"NOT_A_SECRET":        "My app listens on port 3000",
		"OK_SECRET":           "awssm:secret-sauce",
		"BROKEN_SECRET":       "azkv:whisper-alpha.vault.azure.net/does-not-exist",
		"BUGGY_SECRET":        "gcpsm:invalid-ref",
		"NOT_JSON":            "passthrough:not-json|jsonpath:{}",
		"LOOKS_LIKE_A_SECRET": "baz:FAIL",
	}

	_, err := ResolveAll(envVars)
	if err == nil {
		t.Fatal("ResolveAll() returned no error but it should have")
	}

	errMsg := err.Error()

	errorShouldMention := []string{"BROKEN_SECRET", "BUGGY_SECRET"}
	for _, s := range errorShouldMention {
		if !strings.Contains(errMsg, s) {
			t.Errorf("Error message %q should mention %q", errMsg, s)
		}
	}

	errorShouldNotMention := []string{"NOT_A_SECRET", "OK_SECRET", "LOOKS_LIKE_A_SECRET"}
	for _, s := range errorShouldNotMention {
		if strings.Contains(errMsg, s) {
			t.Errorf("Error message %q should not mention %q", errMsg, s)
		}
	}
}
