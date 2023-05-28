package cmd

import (
	"os"

	"github.com/busser/murmur/internal/murmur"
	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a command with secrets injected",
		Long: `Run any command with updated environment variables. Any variables containing
a reference to an externally-stored secret will be overwritten with the secret's
value.

Examples:

  # Azure Key Vault
  export SECRET_SAUCE="azkv:example.vault.azure.net/secret-sauce"
  murmur run -- sh -c 'echo The secret sauce is $SECRET_SAUCE.'`,

		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			exitCode, err := murmur.Run(args[0], args[1:]...)
			if err != nil {
				return err
			}
			os.Exit(exitCode)
			return nil
		},
	}

	return cmd
}
