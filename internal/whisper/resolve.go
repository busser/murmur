package whisper

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"
)

type variable struct {
	// Name of the environment variable.
	name string
	// The environment variable's original value.
	rawValue string
	// The environment variable's query value, if it is a valid whisper query.
	query *query
	// The resolved value of the secret referenced in the query.
	resolvedValue string
	// The filtered value of the secret.
	filteredValue string
	// The final value of the environment variable.
	finalValue string
	// Any error that occurred while processing the environment variable.
	err error
}

// ResolveAll returns a map with the same keys as vars, where all values with
// known prefixes have been replaced with their values.
func ResolveAll(vars map[string]string) (map[string]string, error) {
	var (
		rawVars  = make(chan variable, len(vars))
		parsed   = make(chan variable, len(vars))
		resolved = make(chan variable, len(vars))
		done     = make(chan variable, len(vars))
		failed   = make(chan variable, len(vars))
	)

	// First, feed all the environment variable into the pipeline.

	for name, value := range vars {
		v := variable{
			name:     name,
			rawValue: value,
		}
		rawVars <- v
	}
	close(rawVars)

	// Next, launch the first step of the pipeline: parsing.

	go func() {
		parseVariables(rawVars, parsed, done)
		close(parsed)
	}()

	// Then, launch the second step of the pipeline: reference resolution.

	go func() {
		resolveVariables(parsed, resolved, failed)
		close(resolved)
	}()

	// Next, launch the third step of the pipeline: filtering.

	go func() {
		filterVariables(resolved, done, failed)
		close(done)
		close(failed)
	}()

	// Finally, drain the end of the pipeline and aggregate the results.

	var multierr error
	for v := range failed {
		multierr = multierror.Append(multierr, fmt.Errorf("%s: %w", v.name, v.err))
	}

	if multierr != nil {
		return nil, multierr
	}

	newVars := make(map[string]string)
	for v := range done {
		newVars[v.name] = v.finalValue
	}

	return newVars, nil
}

func parseVariables(rawVars <-chan variable, parsed, done chan<- variable) {
	for v := range rawVars {
		q, err := parseQuery(v.rawValue)
		if err != nil {
			// The variable's value is not a whisper query, so we should leave
			// it as is.
			v.finalValue = v.rawValue
			done <- v
			continue
		}
		if _, known := ProviderFactories[q.providerID]; !known {
			// The variable's value looks like a query but the provider is
			// unknown. It probably isn't a query.
			// ?(busser): should we log a message here?
			v.finalValue = v.rawValue
			done <- v
			continue
		}
		if _, known := Filters[q.filterID]; q.filterID != "" && !known {
			// The variable's value looks like a query but the filter is
			// unknown. It probably isn't a query.
			// ?(busser): should we log a message here?
			v.finalValue = v.rawValue
			done <- v
			continue
		}

		v.query = &q
		parsed <- v
	}
}

// resolveVariables drains `in` and, for each variable, attempts to resolve the
// reference the query contains. Variables with successful resolutions are
// pushed to `out`. Variables with failed resolutions are pushed to `failed`.
func resolveVariables(in <-chan variable, out, failed chan<- variable) {
	chanByProvider := make(map[string]chan variable)
	var wg sync.WaitGroup

	for v := range in {
		providerID := v.query.providerID

		// Dispatch variable to separate goroutines based on the provider
		// required to fetch the secret referenced in the variable's query.
		if _, ok := chanByProvider[providerID]; !ok {
			ch := make(chan variable, cap(in))
			chanByProvider[providerID] = ch

			wg.Add(1)
			go func() {
				resolveVariablesWithProvider(providerID, ch, out, failed)
				wg.Done()
			}()
		}

		chanByProvider[providerID] <- v
	}

	// Wait for each provider to finish resolving its secrets.
	for _, ch := range chanByProvider {
		close(ch)
	}
	wg.Wait()
}

// resolveVariablesWithProvider drains `int` and, for each variable, attempts to
// resolve the reference the query contains with a specific provider. Variables
// with successful resolutions are pushed to `out`. Variables with failed
// resolutions are pushed to `failed`.
func resolveVariablesWithProvider(providerID string, in <-chan variable, out, failed chan<- variable) {
	provider, err := ProviderFactories[providerID]()
	if err != nil {
		// Since we cannot instanciate the provider, we return the same error
		// for all variables sent our way.
		for v := range in {
			v.err = fmt.Errorf("provider instantiation error: %w", err)
			failed <- v
		}
		return
	}
	defer provider.Close()

	// To avoid querying the provider for the same secret twice, we keep a
	// cache of resolved secrets. Since secrets are resolved concurrently,
	// duplicate references are put aside until all unique references have been
	// resolved.

	type result struct {
		secretValue string
		err         error
	}

	var (
		seen       = make(map[string]bool)
		duplicates []variable

		wg sync.WaitGroup

		mu    sync.Mutex // protects cache
		cache = make(map[string]result)
	)

	for v := range in {
		if seen[v.query.secretRef] {
			duplicates = append(duplicates, v)
			continue
		}
		seen[v.query.secretRef] = true

		wg.Add(1)
		go func(v variable) {
			defer wg.Done()

			secretValue, err := provider.Resolve(context.TODO(), v.query.secretRef)

			mu.Lock()
			cache[v.query.secretRef] = result{secretValue, err}
			mu.Unlock()

			if err != nil {
				v.err = fmt.Errorf("could not resolve reference: %w", err)
				failed <- v
				return
			}

			v.resolvedValue = secretValue
			out <- v
		}(v)
	}

	wg.Wait()

	// Now that all unique references have been resolved, results for duplicate
	// references can be read from cache.

	for _, v := range duplicates {
		result := cache[v.query.secretRef]
		if result.err != nil {
			v.err = fmt.Errorf("could not resolve reference: %w", err)
			failed <- v
			continue
		}

		v.resolvedValue = result.secretValue
		out <- v
	}
}

// filterVariables drains `in` and, for each variable, attempts to filter the
// the secret's value with the filtering rule contained in the query. Variables
// with successful resolutions are pushed to `out`. Variables with failed
// resolutions are pushed to `failed`.
func filterVariables(in <-chan variable, out, failed chan<- variable) {
	chanByFilter := make(map[string]chan variable)
	var wg sync.WaitGroup

	for v := range in {
		filterID := v.query.filterID

		if filterID == "" {
			v.filteredValue = v.resolvedValue
			v.finalValue = v.filteredValue
			out <- v
			continue
		}

		// Dispatch variable to separate goroutines based on the filter
		// specified in the variable's query.
		if _, ok := chanByFilter[filterID]; !ok {
			ch := make(chan variable, cap(in))
			chanByFilter[filterID] = ch

			wg.Add(1)
			go func() {
				filterVariablesWithFilter(filterID, ch, out, failed)
				wg.Done()
			}()
		}

		chanByFilter[filterID] <- v
	}

	// Wait for each filter to finish filtering its secrets.
	for _, ch := range chanByFilter {
		close(ch)
	}
	wg.Wait()
}

// filterVariablesWithFilter drains `in` and, for each variable, attempts to
// filter the the secret's value with a specific filter. Variables with
// successful resolutions are pushed to `out`. Variables with failed resolutions
// are pushed to `failed`.
func filterVariablesWithFilter(filterID string, in <-chan variable, out, failed chan<- variable) {
	filter := Filters[filterID]

	for v := range in {
		filteredValue, err := filter(v.resolvedValue, v.query.filterRule)
		if err != nil {
			v.err = fmt.Errorf("could not filter value: %w", err)
			failed <- v
			continue
		}

		v.filteredValue = filteredValue
		v.finalValue = v.filteredValue
		out <- v
	}
}
