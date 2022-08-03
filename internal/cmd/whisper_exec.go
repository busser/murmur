package cmd

import (
	"os"

	"github.com/busser/whisper/internal/whisper"
	"github.com/spf13/cobra"
)

func execCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec",
		Short: "Execute a command with secrets injected",
		Long: `Execute any command with updated environment variables. Any variables containing
a reference to an externally-stored secret will be overwritten with the secret's
value.

Examples:

  # Azure Key Vault
  export WHISPER_AZURE_KEY_VAULT_URL="https://example.vault.azure.net/"
  export SECRET_SAUCE="azkv:secret-sauce"
  whisper exec -- sh -c 'echo The secret sauce is $SECRET_SAUCE.'`,

		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			exitCode, err := whisper.Exec(args[0], args[1:]...)
			if err != nil {
				return err
			}
			os.Exit(exitCode)
			return nil
		},
	}

	return cmd
}
