package cli

import (
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/spf13/cobra"
)

var cmdInfo = &cobra.Command{
	Use:   "info",
	Short: "",
	RunE:  runInfo,
}

func init() {
	cmdRoot.AddCommand(cmdInfo)
}

func runInfo(cmd *cobra.Command, args []string) error {
	return steps.InfoOptional(dock, deb, n)
}
