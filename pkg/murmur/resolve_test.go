package murmur

import (
	"strings"
	"testing"

	"github.com/busser/murmur/pkg/murmur/providers/jsonmock"
	"github.com/busser/murmur/pkg/murmur/providers/mock"
	"github.com/busser/murmur/pkg/slices"
	"github.com/google/go-cmp/cmp"
)

type MockProvider interface {
	Provider
	ResolvedRefs() []string
	Closed() bool
}

func TestResolveAll(t *testing.T) {
	tt := []struct {
		name      string
		providers map[string]MockProvider
		variables map[string]string
		want      map[string]string
	}{
		{
			name: "no overloads",
			variables: map[string]string{
				"A": "A",
				"B": "B",
				"C": "bar:C",
			},
			want: map[string]string{
				"A": "A",
				"B": "B",
				"C": "bar:C",
			},
		},
		{
			name: "multiple providers",
			providers: map[string]MockProvider{
				"foo":  mock.New(),
				"bar":  mock.New(),
				"json": jsonmock.New(),
			},
			variables: map[string]string{
				"A": "foo:A",
				"B": "foo:B",
				"C": "bar:C",
				"D": "json:D",
			},
			want: map[string]string{
				"A": mock.ValueFor("A"),
				"B": mock.ValueFor("B"),
				"C": mock.ValueFor("C"),
				"D": jsonmock.ValueFor("D"),
			},
		},
		{
			name: "filters",
			providers: map[string]MockProvider{
				"json": jsonmock.New(),
			},
			variables: map[string]string{
				"A": "json:A|jsonpath:{ ." + jsonmock.Key + " }",
				"B": "json:B|jsonpath:ref={ ." + jsonmock.Key + " }",
				"C": "json:C|jsonpath:is my ref { ." + jsonmock.Key + " }?",
			},
			want: map[string]string{
				"A": "A",
				"B": "ref=B",
				"C": "is my ref C?",
			},
		},
		{
			name: "caching",
			providers: map[string]MockProvider{
				"json": jsonmock.New(),
			},
			variables: map[string]string{
				"A": "json:A|jsonpath:{ ." + jsonmock.Key + " }",
				"B": "json:A|jsonpath:ref={ ." + jsonmock.Key + " }",
				"C": "json:A|jsonpath:is my ref { ." + jsonmock.Key + " }?",
			},
			want: map[string]string{
				"A": "A",
				"B": "ref=A",
				"C": "is my ref A?",
			},
		},
		{
			name: "a bit of everything",
			providers: map[string]MockProvider{
				"foo":  mock.New(),
				"bar":  mock.New(),
				"json": jsonmock.New(),
			},
			variables: map[string]string{
				"NOT_A_SECRET":        "My app listens on port 3000",
				"NOT_A_SECRET_EITHER": "The cloud is awesome",
				"FIRST_SECRET":        "foo:database password",
				"SECOND_SECRET":       "foo:private key",
				"THIRD_SECRET":        "bar:api key",
				"FOURTH_SECRET":       "bar:api key",
				"LOOKS_LIKE_A_SECRET": "baz:but isn't a secret",
				"JSON_SECRET":         "json:cloud credentials|jsonpath:{ ." + jsonmock.Key + " }",
				"SAME_JSON_SECRET":    "json:cloud credentials|jsonpath:ref={ ." + jsonmock.Key + " }",
			},
			want: map[string]string{
				"NOT_A_SECRET":        "My app listens on port 3000",
				"NOT_A_SECRET_EITHER": "The cloud is awesome",
				"FIRST_SECRET":        mock.ValueFor("database password"),
				"SECOND_SECRET":       mock.ValueFor("private key"),
				"THIRD_SECRET":        mock.ValueFor("api key"),
				"FOURTH_SECRET":       mock.ValueFor("api key"),
				"LOOKS_LIKE_A_SECRET": "baz:but isn't a secret",
				"JSON_SECRET":         "cloud credentials",
				"SAME_JSON_SECRET":    "ref=cloud credentials",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			factories := make(map[string]ProviderFactory)
			for prefix, provider := range tc.providers {
				provider := provider
				factories[prefix] = func() (Provider, error) { return provider, nil }
			}

			// Replace murmur's clients with mocks for the duration of the test.
			originalProviderFactories := ProviderFactories
			defer func() { ProviderFactories = originalProviderFactories }()
			ProviderFactories = factories

			actual, err := ResolveAll(tc.variables)
			if err != nil {
				t.Fatalf("ResolveAll() returned an error: %v", err)
			}

			for prefix, provider := range tc.providers {
				if !provider.Closed() {
					t.Errorf("%q provider not closed", prefix)
				}
				if slices.Duplicates(provider.ResolvedRefs()) != 0 {
					t.Errorf("%q provider resolved the same reference more than once, is caching broken?", prefix)
					t.Logf("%q provider resolved: %q", prefix, provider.ResolvedRefs())
				}
			}

			if diff := cmp.Diff(tc.want, actual); diff != "" {
				t.Errorf("ResolveAll() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResolveSecrets(t *testing.T) {
	tt := []struct {
		name      string
		providers map[string]MockProvider
		variables map[string]string
		want      map[string]string
	}{
		{
			name: "no secrets",
			variables: map[string]string{
				"A": "A",
				"B": "B",
				"C": "bar:C", // looks like query but provider unknown
			},
			want: map[string]string{}, // empty result - no valid secrets
		},
		{
			name: "mixed secrets and non-secrets",
			providers: map[string]MockProvider{
				"foo":  mock.New(),
				"bar":  mock.New(),
				"json": jsonmock.New(),
			},
			variables: map[string]string{
				"NOT_A_SECRET":   "My app listens on port 3000",
				"ALSO_NOT":       "The cloud is awesome",
				"SECRET_ONE":     "foo:database password",
				"SECRET_TWO":     "bar:api key",
				"JSON_SECRET":    "json:cloud credentials|jsonpath:{ ." + jsonmock.Key + " }",
				"LOOKS_LIKE_ONE": "baz:but isn't a secret", // unknown provider
			},
			want: map[string]string{
				"SECRET_ONE":  mock.ValueFor("database password"),
				"SECRET_TWO":  mock.ValueFor("api key"),
				"JSON_SECRET": "cloud credentials",
			},
		},
		{
			name: "only secrets",
			providers: map[string]MockProvider{
				"foo":  mock.New(),
				"json": jsonmock.New(),
			},
			variables: map[string]string{
				"A": "foo:A",
				"B": "foo:B",
				"C": "json:C|jsonpath:{ ." + jsonmock.Key + " }",
			},
			want: map[string]string{
				"A": mock.ValueFor("A"),
				"B": mock.ValueFor("B"),
				"C": "C",
			},
		},
		{
			name: "caching with secrets only",
			providers: map[string]MockProvider{
				"json": jsonmock.New(),
			},
			variables: map[string]string{
				"NOT_SECRET": "plain value",
				"A":          "json:A|jsonpath:{ ." + jsonmock.Key + " }",
				"B":          "json:A|jsonpath:ref={ ." + jsonmock.Key + " }",
				"C":          "json:A|jsonpath:is my ref { ." + jsonmock.Key + " }?",
			},
			want: map[string]string{
				"A": "A",
				"B": "ref=A",
				"C": "is my ref A?",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			factories := make(map[string]ProviderFactory)
			for prefix, provider := range tc.providers {
				provider := provider
				factories[prefix] = func() (Provider, error) { return provider, nil }
			}

			// Replace murmur's clients with mocks for the duration of the test.
			originalProviderFactories := ProviderFactories
			defer func() { ProviderFactories = originalProviderFactories }()
			ProviderFactories = factories

			actual, err := ResolveSecrets(tc.variables)
			if err != nil {
				t.Fatalf("ResolveSecrets() returned an error: %v", err)
			}

			for prefix, provider := range tc.providers {
				if !provider.Closed() {
					t.Errorf("%q provider not closed", prefix)
				}
				if slices.Duplicates(provider.ResolvedRefs()) != 0 {
					t.Errorf("%q provider resolved the same reference more than once, is caching broken?", prefix)
					t.Logf("%q provider resolved: %q", prefix, provider.ResolvedRefs())
				}
			}

			if diff := cmp.Diff(tc.want, actual); diff != "" {
				t.Errorf("ResolveSecrets() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResolveSecretsWithError(t *testing.T) {
	tt := []struct {
		name       string
		providers  map[string]MockProvider
		variables  map[string]string
		wantFailed []string
	}{
		{
			name: "mixed with errors",
			providers: map[string]MockProvider{
				"foo":  mock.New(),
				"json": jsonmock.New(),
			},
			variables: map[string]string{
				"NOT_A_SECRET":   "My app listens on port 3000",
				"OK_SECRET":      "foo:database password",
				"BROKEN_SECRET":  "foo:FAIL",
				"JSON_ERR":       "json:cloud credentials|jsonpath:{ .missing }",
				"NOT_JSON":       "foo:api key|jsonpath:{ .foo }",
				"LOOKS_LIKE_ONE": "baz:FAIL", // unknown provider, should be ignored
			},
			wantFailed: []string{"BROKEN_SECRET", "JSON_ERR", "NOT_JSON"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			factories := make(map[string]ProviderFactory)
			for prefix, provider := range tc.providers {
				provider := provider
				factories[prefix] = func() (Provider, error) { return provider, nil }
			}

			// Replace murmur's clients with mocks for the duration of the test.
			originalProviderFactories := ProviderFactories
			defer func() { ProviderFactories = originalProviderFactories }()
			ProviderFactories = factories

			_, err := ResolveSecrets(tc.variables)
			if err == nil {
				t.Fatal("ResolveSecrets() returned no error but it should have")
			}

			for prefix, provider := range tc.providers {
				if !provider.Closed() {
					t.Errorf("%q provider not closed", prefix)
				}
				if slices.Duplicates(provider.ResolvedRefs()) != 0 {
					t.Errorf("%q provider resolved the same reference more than once, is caching broken?", prefix)
					t.Logf("%q provider resolved: %q", prefix, provider.ResolvedRefs())
				}
			}

			errMsg := err.Error()

			for _, s := range tc.wantFailed {
				if !strings.Contains(errMsg, s) {
					t.Errorf("Error message %q should mention %q", errMsg, s)
				}
			}
		})
	}
}

func TestResolveAllWithError(t *testing.T) {
	tt := []struct {
		name       string
		providers  map[string]MockProvider
		variables  map[string]string
		wantOK     []string
		wantFailed []string
	}{
		{
			name: "a bit of everything",
			providers: map[string]MockProvider{
				"foo":  mock.New(),
				"bar":  mock.New(),
				"json": jsonmock.New(),
			},
			variables: map[string]string{
				"NOT_A_SECRET":        "My app listens on port 3000",
				"OK_SECRET":           "foo:database password",
				"BROKEN_SECRET":       "foo:FAIL",
				"BUGGY_SECRET":        "bar:FAIL",
				"LOOKS_LIKE_A_SECRET": "baz:FAIL",
				"JSON_ERR":            "json:cloud credentials|jsonpath:{ .missing }",
				"NOT_JSON":            "foo:api key|jsonpath:{ .foo }",
				"OK_JSON":             "json:cloud credentials|jsonpath:{ ." + jsonmock.Key + " }",
			},
			wantOK:     []string{"NOT_A_SECRET", "OK_SECRET", "LOOKS_LIKE_A_SECRET", "OK_JSON"},
			wantFailed: []string{"BROKEN_SECRET", "BUGGY_SECRET", "JSON_ERR", "NOT_JSON"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			factories := make(map[string]ProviderFactory)
			for prefix, provider := range tc.providers {
				provider := provider
				factories[prefix] = func() (Provider, error) { return provider, nil }
			}

			// Replace murmur's clients with mocks for the duration of the test.
			originalProviderFactories := ProviderFactories
			defer func() { ProviderFactories = originalProviderFactories }()
			ProviderFactories = factories

			_, err := ResolveAll(tc.variables)
			if err == nil {
				t.Fatal("ResolveAll() returned no error but it should have")
			}

			for prefix, provider := range tc.providers {
				if !provider.Closed() {
					t.Errorf("%q provider not closed", prefix)
				}
				if slices.Duplicates(provider.ResolvedRefs()) != 0 {
					t.Errorf("%q provider resolved the same reference more than once, is caching broken?", prefix)
					t.Logf("%q provider resolved: %q", prefix, provider.ResolvedRefs())
				}
			}

			errMsg := err.Error()

			for _, s := range tc.wantOK {
				if strings.Contains(errMsg, s) {
					t.Errorf("Error message %q should not mention %q", errMsg, s)
				}
			}

			for _, s := range tc.wantFailed {
				if !strings.Contains(errMsg, s) {
					t.Errorf("Error message %q should mention %q", errMsg, s)
				}
			}
		})
	}
}
