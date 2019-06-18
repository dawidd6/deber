package cli

import (
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/spf13/cobra"
)

var cmdShell = &cobra.Command{
	Use:   "shell",
	Short: "",
	RunE:  runShell,
}

func init() {
	cmdRoot.AddCommand(cmdShell)
}

func runShell(cmd *cobra.Command, args []string) error {
	return steps.ShellOptional(dock, deb, n)
}
