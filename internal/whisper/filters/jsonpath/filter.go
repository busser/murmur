package jsonpath

import (
	"bytes"
	"encoding/json"
	"errors"
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
		return "", errors.New("value is not valid JSON")
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, parsedValue); err != nil {
		return "", fmt.Errorf("could not render template: %w", err)
	}

	return buf.String(), nil
}
