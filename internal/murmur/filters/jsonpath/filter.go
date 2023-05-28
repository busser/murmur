package jsonpath

import (
	"bytes"
	"encoding/json"
	"fmt"

	"k8s.io/client-go/util/jsonpath"
)

// Filter renders the given template with the given value. Filter returns an
// error if the value is not valid JSON or if the template is invalid.
// This function uses Kubernetes' JSONPath syntax as documented here:
// https://kubernetes.io/docs/reference/kubectl/jsonpath/.
func Filter(value, template string) (string, error) {
	tmpl := jsonpath.New("filter")

	if err := tmpl.Parse(template); err != nil {
		return "", fmt.Errorf("invalid jsonpath template: %w", err)
	}

	var parsedValue any
	if err := json.Unmarshal([]byte(value), &parsedValue); err != nil {
		// If the value is not valid JSON, we can still use it as a string.
		parsedValue = value
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, parsedValue); err != nil {
		return "", fmt.Errorf("could not render template: %w", err)
	}

	return buf.String(), nil
}
