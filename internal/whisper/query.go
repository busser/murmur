package whisper

import (
	"errors"
	"fmt"
	"strings"
)

type query struct {
	providerID string
	secretRef  string

	filterID   string
	filterRule string
}

func parseQuery(s string) (query, error) {
	if len(s) == 0 {
		return query{}, errors.New("empty query")
	}

	const separator = "|"

	parts := strings.SplitN(s, separator, 2)

	providerID, secretRef, err := parseQuerySecret(parts[0])
	if err != nil {
		return query{}, fmt.Errorf("left of first %q: %w", separator, err)
	}

	if len(parts) == 1 {
		return query{
			providerID: providerID,
			secretRef:  secretRef,
		}, nil
	}

	filterID, filterRule, err := parseQueryFilter(parts[1])
	if err != nil {
		return query{}, fmt.Errorf("right of first %q: %w", separator, err)
	}

	return query{
		providerID: providerID,
		secretRef:  secretRef,
		filterID:   filterID,
		filterRule: filterRule,
	}, nil
}

func parseQuerySecret(s string) (providerID, secretRef string, err error) {
	const separator = ":"

	parts := strings.SplitN(s, separator, 2)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("must be at least two parts separated by %q", separator)
	}

	if parts[0] == "" {
		return "", "", fmt.Errorf("provider ID left of %q cannot be empty string", separator)
	}

	if parts[1] == "" {
		return "", "", fmt.Errorf("reference right of %q cannot be empty string", separator)
	}

	return parts[0], parts[1], nil
}

func parseQueryFilter(s string) (filterID, filterRule string, err error) {
	const separator = ":"

	parts := strings.SplitN(s, separator, 2)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("must be at least two parts separated by %q", separator)
	}

	if parts[0] == "" {
		return "", "", fmt.Errorf("filter ID left of %q cannot be empty string", separator)
	}

	if parts[1] == "" {
		return "", "", fmt.Errorf("filter rule right of %q cannot be empty string", separator)
	}

	return parts[0], parts[1], nil
}
