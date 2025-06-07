package cmd

import (
	"github.com/spf13/cobra"
)

func execCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "exec -- command [args...]",
		Args: cobra.MinimumNArgs(1),

		DisableFlagsInUseLine: true,

		Deprecated: "command \"run\" has the same behavior.",

		RunE: runCmd().RunE,
	}

	return cmd
}
