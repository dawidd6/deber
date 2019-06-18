package cli

import (
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/spf13/cobra"
)

var (
	dist string
)

var cmdBuild = &cobra.Command{
	Use:   "build",
	Short: "",
	RunE:  runBuild,
}

func init() {
	cmdRoot.AddCommand(cmdBuild)

	cmdBuild.Flags().BoolVarP(&steps.NoRebuild, "no-rebuild", "n", steps.NoRebuild, "")
	cmdBuild.Flags().StringVarP(&dist, "distribution", "d", dist, "")
}

func runBuild(cmd *cobra.Command, args []string) error {
	return steps.Build(dock, deb, n)
}
