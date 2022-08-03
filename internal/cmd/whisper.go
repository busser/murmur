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
		Use:   "whisper",
		Short: "Whisper injects secrets into your application",
		Long: `A plug-and-play entrypoint that fetches secrets from a secure
	location and adds them to your application's environment variables.`,
	}

	cmd.AddCommand(execCmd())

	return cmd
}
