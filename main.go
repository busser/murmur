package main

import (
	"os"

	extension "github.com/busser/murmur/pkg/aws/lambda-extension"
	"github.com/busser/murmur/pkg/cmd"
)

func main() {
	// Detect if running in Lambda environment by checking for AWS_LAMBDA_RUNTIME_API
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" {
		// Lambda extension mode
		extension.Execute()
	} else {
		// CLI mode
		cmd.Execute()
	}
}
