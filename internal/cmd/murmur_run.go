package cmd

import (
	"os"

	"github.com/busser/murmur/internal/murmur"
	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	var verbose *bool

	cmd := &cobra.Command{
		Use:  "run -- command [args...]",
		Args: cobra.MinimumNArgs(1),

		DisableFlagsInUseLine: false,

		Short: "Run a command with secrets injected into its environment variables",

		Example: `  # Fetch a database password from Scaleway Secret Manager:
  export PGPASSWORD="scwsm:database-password"
  murmur run -- psql -h 10.1.12.34 -U my-user -d my-database  
  
  # Build a connection string from a JSON secret:
  export PGDATABASE="scwsm:database-credentials|jsonpath:{.username}:{password}@{.host}:{.port}/{.database}" 
  murmur run -- psql`,

		RunE: func(cmd *cobra.Command, args []string) error {
			exitCode, err := murmur.Run(*verbose, args[0], args[1:]...)
			if err != nil {
				return err
			}
			os.Exit(exitCode)
			return nil
		},
	}
	verbose = cmd.Flags().BoolP("verbose", "v", false, "enables extra logging")

	return cmd
}
