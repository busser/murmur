package environ

import (
	"os"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestToMap(t *testing.T) {
	tt := []struct {
		env  []string
		want map[string]string
	}{
		{
			env:  []string{"a=b", "c=d"},
			want: map[string]string{"a": "b", "c": "d"},
		},
	}

	for _, tc := range tt {
		actual := ToMap(tc.env)
		if diff := cmp.Diff(tc.want, actual); diff != "" {
			t.Errorf("ToMap() mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestToSlice(t *testing.T) {
	tt := []struct {
		env  map[string]string
		want []string
	}{
		{
			env:  map[string]string{"a": "b", "c": "d"},
			want: []string{"a=b", "c=d"},
		},
	}

	for _, tc := range tt {
		actual := ToSlice(tc.env)
		sort.Strings(tc.want)
		sort.Strings(actual)
		if diff := cmp.Diff(tc.want, actual); diff != "" {
			t.Errorf("ToMap() mismatch (-want +got):\n%s", diff)
		}
	}
}
func TestGetEnvWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		envVar       string
		envValue     string
		defaultValue string
		want         string
	}{
		{
			name:         "environment variable set",
			envVar:       "TEST_VAR",
			envValue:     "env_value",
			defaultValue: "default_value",
			want:         "env_value",
		},
		{
			name:         "environment variable not set",
			envVar:       "UNSET_VAR",
			envValue:     "",
			defaultValue: "default_value",
			want:         "default_value",
		},
		{
			name:         "environment variable empty",
			envVar:       "EMPTY_VAR",
			envValue:     "",
			defaultValue: "default_value",
			want:         "default_value",
		},
		{
			name:         "both empty",
			envVar:       "UNSET_VAR",
			envValue:     "",
			defaultValue: "",
			want:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment variable
			os.Unsetenv(tt.envVar)

			// Set environment variable if needed
			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
			}

			got := GetEnvWithDefault(tt.envVar, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetEnvWithDefault() = %v, want %v", got, tt.want)
			}

			// Clean up
			os.Unsetenv(tt.envVar)
		})
	}
}
