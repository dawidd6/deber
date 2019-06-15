package cli

import (
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
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
	dock, err := docker.New()
	if err != nil {
		return err
	}

	deb := new(debian.Debian)

	if dist == "" {
		deb, err = debian.New()
		if err != nil {
			return err
		}
	} else {
		deb.Target = dist
	}

	n := naming.New(deb)

	return steps.Build(dock, deb, n)
}
