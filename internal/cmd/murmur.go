package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func Execute() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "murmur",
		Short: "Murmur passes secrets as environment variables to a process",
		Long: `A plug-and-play shim that fetches secrets from a secure
	location and passes them to your application as environment variables.`,
		SilenceUsage: true,
	}

	cmd.AddCommand(runCmd())

	return cmd
}
