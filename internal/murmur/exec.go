package murmur

import (
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"sort"

	"github.com/busser/murmur/internal/environ"
)

// Modified during testing to catch command output.
var (
	runOut io.Writer = os.Stdout
	runErr io.Writer = os.Stderr
)

func Run(name string, args ...string) (exitCode int, err error) {
	originalVars := environ.ToMap(os.Environ())

	newVars, err := ResolveAll(originalVars)
	if err != nil {
		return 0, err
	}

	var overloaded []string
	for name, original := range originalVars {
		if newVars[name] != original {
			overloaded = append(overloaded, name)
		}
	}

	sort.Strings(overloaded)
	for _, name := range overloaded {
		log.Printf("[murmur] overloading %s", name)
	}

	subCmd := exec.Command(name, args...)
	subCmd.Env = environ.ToSlice(newVars)
	subCmd.Stdin = os.Stdin
	subCmd.Stdout = runOut
	subCmd.Stderr = runErr

	if err := subCmd.Run(); err != nil {
		exitErr := new(exec.ExitError)
		if errors.As(err, &exitErr) {
			return exitErr.ProcessState.ExitCode(), nil
		}
		return 0, err
	}

	return subCmd.ProcessState.ExitCode(), nil
}
