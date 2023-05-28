package murmur

import (
	"reflect"
	"testing"
)

func TestParseQuery(t *testing.T) {
	tt := []struct {
		s       string
		want    query
		wantErr bool
	}{
		{
			s: "my_provider:my_secret",
			want: query{
				providerID: "my_provider",
				secretRef:  "my_secret",
			},
		},
		{
			s: "my_provider:my_secret:my_version",
			want: query{
				providerID: "my_provider",
				secretRef:  "my_secret:my_version",
			},
		},
		{
			s: "my_provider:my_secret:my_version|my_filter:my_filter_rule",
			want: query{
				providerID: "my_provider",
				secretRef:  "my_secret:my_version",
				filterID:   "my_filter",
				filterRule: "my_filter_rule",
			},
		},
		{
			s: "my_provider:my_secret:my_version|my_filter:my:complex|filter:rule",
			want: query{
				providerID: "my_provider",
				secretRef:  "my_secret:my_version",
				filterID:   "my_filter",
				filterRule: "my:complex|filter:rule",
			},
		},
		{
			s:       "",
			wantErr: true,
		},
		{
			s:       "my_provider",
			wantErr: true,
		},
		{
			s:       "my_provider:",
			wantErr: true,
		},
		{
			s:       ":my_secret",
			wantErr: true,
		},
		{
			s:       ":my_secret:my_version",
			wantErr: true,
		},
		{
			s:       "my_provider:my_secret|",
			wantErr: true,
		},
		{
			s:       "my_provider:my_secret|my_filter",
			wantErr: true,
		},
		{
			s:       "my_provider:my_secret|my_filter:",
			wantErr: true,
		},
		{
			s:       "my_provider:my_secret|:my_filter_rule",
			wantErr: true,
		},
		{
			s:       "my_provider:my_secret|:my:complex:filter:rule",
			wantErr: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.s, func(t *testing.T) {
			actual, err := parseQuery(tc.s)

			if err != nil && !tc.wantErr {
				t.Errorf("unexpected error: %v", err)
			}
			if err == nil && tc.wantErr {
				t.Errorf("expected error, got none")
			}

			if !reflect.DeepEqual(actual, tc.want) {
				t.Errorf("ParseQuery() = %#v, want %#v", actual, tc.want)
			}
		})
	}
}
