package murmur

import (
	"bytes"
	"os"
	"testing"

	"github.com/busser/murmur/pkg/environ"
)

func TestRun(t *testing.T) {
	tt := []struct {
		name         string
		command      []string
		env          []string
		wantExitCode int
		wantOutput   string
	}{
		{
			name:         "exit code 0",
			command:      []string{"/bin/sh", "-c", "exit 0"},
			env:          nil,
			wantExitCode: 0,
			wantOutput:   "",
		},
		{
			name:         "exit code 1",
			command:      []string{"/bin/sh", "-c", "exit 1"},
			env:          nil,
			wantExitCode: 1,
			wantOutput:   "",
		},
		{
			name:         "exit code 123",
			command:      []string{"/bin/sh", "-c", "exit 123"},
			env:          nil,
			wantExitCode: 123,
			wantOutput:   "",
		},
		{
			name:         "env without replacement",
			command:      []string{"/bin/sh", "-c", "printenv SECRET_SAUCE"},
			env:          []string{"SECRET_SAUCE=szechuan"},
			wantExitCode: 0,
			wantOutput:   "szechuan\n",
		},
		{
			name:         "env with replacement",
			command:      []string{"/bin/sh", "-c", "printenv SECRET_SAUCE"},
			env:          []string{"SECRET_SAUCE=passthrough:szechuan"},
			wantExitCode: 0,
			wantOutput:   "szechuan\n",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			// Capture Run()'s output for the duration of the test.
			var output bytes.Buffer
			runOut = &output
			runErr = &output
			defer func() {
				runOut = os.Stdout
				runErr = os.Stderr
			}()

			// Clear all environment variables for the duration of the test.
			originalEnv := os.Environ()
			os.Clearenv()
			defer func() {
				os.Clearenv()
				for k, v := range environ.ToMap(originalEnv) {
					os.Setenv(k, v)
				}
			}()

			// Set specific environment variables for this test.
			for k, v := range environ.ToMap(tc.env) {
				os.Setenv(k, v)
			}

			exitCode, err := Run(tc.command[0], tc.command[1:]...)
			if err != nil {
				t.Errorf("Run() returned an error: %v", err)
			}

			if exitCode != tc.wantExitCode {
				t.Errorf("got exit code %d, want %d", exitCode, tc.wantExitCode)
			}

			if output.String() != tc.wantOutput {
				t.Errorf("got output %q, want %q", output.String(), tc.wantOutput)
			}
		})
	}
}
