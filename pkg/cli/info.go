package cli

import (
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
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
	dock, err := docker.New()
	if err != nil {
		return err
	}

	deb, err := debian.New()
	if err != nil {
		return err
	}

	n := naming.New(deb)

	return steps.InfoOptional(dock, deb, n)
}
