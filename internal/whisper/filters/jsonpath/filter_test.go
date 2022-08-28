package jsonpath

import (
	"testing"
)

func TestFilter(t *testing.T) {
	tt := []struct {
		name     string
		value    string
		template string
		want     string
		wantErr  bool
	}{
		{
			name:     "static template",
			value:    `{"foo": "bar"}`,
			template: "hello",
			want:     "hello",
		},
		{
			name:     "single value",
			value:    `{"foo": "bar"}`,
			template: "foo={ .foo }",
			want:     "foo=bar",
		},
		{
			name:     "nested value",
			value:    `{"foo": {"bar": "baz"}}`,
			template: "foobar={ .foo.bar }",
			want:     "foobar=baz",
		},
		{
			name:     "missing value",
			value:    `{"foo": "bar"}`,
			template: "{ .missing }",
			wantErr:  true,
		},
		{
			name:     "empty template",
			value:    `{"foo": "bar"}`,
			template: "",
			want:     "",
		},
		{
			name:     "invalid json",
			value:    "foo",
			template: "",
			wantErr:  true,
		},
		{
			name:     "invalid template",
			value:    `{"foo": "bar"}`,
			template: "{ .not_closed",
			wantErr:  true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			actual, err := Filter(tc.value, tc.template)

			if err != nil && !tc.wantErr {
				t.Errorf("Filter() returned an error: %v", err)
			}
			if err == nil && tc.wantErr {
				t.Error("Filter() did not return an error")
			}

			if actual != tc.want {
				t.Errorf("Filter() returned %q, want %q", actual, tc.want)
			}
		})
	}
}
