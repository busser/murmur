package murmur

import (
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
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

	if err := subCmd.Start(); err != nil {
		return 1, err
	}

	// Capture signals for the duration of the sub process and forward them.
	// This is challenging to test automatically, so it's not.
	// This feature can be tested manually by running this command:
	// murmur run -- ./internal/murmur/testdata/signal.sh
	// and then sending an interrupt signal to the murmur process.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals)

	stop := make(chan struct{})
	defer func() {
		close(stop)
	}()

	// Forward signals to the sub process.
	go func() {
		for {
			select {
			case <-stop:
				return
			case sig := <-signals:
				_ = subCmd.Process.Signal(sig)
			}
		}
	}()

	if err := subCmd.Wait(); err != nil {
		exitErr := new(exec.ExitError)
		if errors.As(err, &exitErr) {
			return exitErr.ProcessState.ExitCode(), nil
		}
		return 0, err
	}

	return subCmd.ProcessState.ExitCode(), nil
}
