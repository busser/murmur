package environ_test

import (
	"fmt"
	"os"

	"github.com/busser/whisper/internal/environ"
)

func Example() {
	os.Setenv("foo", "bar")

	envMap := environ.ToMap(os.Environ())

	fmt.Println(envMap["foo"])

	// Output:
	// bar
}
