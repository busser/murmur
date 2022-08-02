package whisper

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestExec(t *testing.T) {
	tt := []struct {
		name             string
		command          string
		env              []string
		expectedExitCode int
		expectedOutput   string
	}{
		{
			"exit code 0",
			"exit 0",
			nil,
			0,
			"",
		},
		{
			"exit code 1",
			"exit 1",
			nil,
			1,
			"",
		},
		{
			"exit code 123",
			"exit 123",
			nil,
			123,
			"",
		},
		{
			"env without replacement",
			"printenv SECRET_SAUCE",
			[]string{"SECRET_SAUCE=szechuan"},
			0,
			"szechuan\n",
		},
		{
			"env with replacement",
			"printenv SECRET_SAUCE",
			[]string{"SECRET_SAUCE=passthrough:szechuan"},
			0,
			"szechuan\n",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			args := []string{"-test.run=TestExecHelperProcess", "--"}
			args = append(args, "/bin/sh", "-c", tc.command)

			cmd := exec.Command(os.Args[0], args...)
			cmd.Env = append(tc.env, "GO_WANT_HELPER_PROCESS=1")

			var execStdout, execStderr bytes.Buffer
			cmd.Stdout = &execStdout
			cmd.Stderr = &execStderr

			if err := cmd.Run(); err != nil {
				exitErr := new(exec.ExitError)
				if !errors.As(err, &exitErr) {
					t.Fatalf("failed to run Exec() is subprocess: %v", err)
				}
				if exitErr.ProcessState.ExitCode() == helperProcessErrorCode {
					t.Fatalf("the helper process encountered an error: %s", execStderr.String())
				}
			}

			exitCode := cmd.ProcessState.ExitCode()
			if exitCode != tc.expectedExitCode {
				t.Errorf("got exit code %d, want %d", exitCode, tc.expectedExitCode)
			}

			if execStdout.String() != tc.expectedOutput {
				t.Errorf("got output %q, want %q", execStdout.String(), tc.expectedOutput)
			}
		})
	}
}

const helperProcessErrorCode = 3

func TestExecHelperProcess(t *testing.T) {

	// This test does not actually test anything. Instead, it allows TestExec()
	// to run the Exec function in a separate process. This enables validation
	// of Exec's exit code and output.

	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}

		args = args[1:]
	}

	exitCode, err := Exec(args[0], args[1:]...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Exec() returned an error: %v", err)
		os.Exit(helperProcessErrorCode)
	}

	os.Exit(exitCode)
}
