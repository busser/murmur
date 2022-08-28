package whisper

import "github.com/busser/whisper/internal/whisper/filters/jsonpath"

// A Filter transforms a value obtained from a secret store into another value
// based on the given rule.
type Filter func(value, rule string) (string, error)

var Filters = map[string]Filter{
	// Kubernetes JSONPath templating.
	"jsonpath": jsonpath.Filter,
}
