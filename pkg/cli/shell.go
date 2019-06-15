package cli

import (
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
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
	dock, err := docker.New()
	if err != nil {
		return err
	}

	deb, err := debian.New()
	if err != nil {
		return err
	}

	n := naming.New(deb)

	return steps.ShellOptional(dock, deb, n)
}
