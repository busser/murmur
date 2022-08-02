package environ

import (
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
