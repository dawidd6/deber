package cli

import (
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/spf13/cobra"
)

var (
	start bool
)

var cmdCreate = &cobra.Command{
	Use:   "create",
	Short: "",
	RunE:  runCreate,
}

func init() {
	cmdRoot.AddCommand(cmdCreate)

	cmdCreate.Flags().BoolVarP(&start, "start", "s", start, "")
	cmdCreate.Flags().StringArrayVar(&steps.ExtraPackages, "extra-package", steps.ExtraPackages, "")
}

func runCreate(cmd *cobra.Command, args []string) error {
	dock, err := docker.New()
	if err != nil {
		return err
	}

	deb, err := debian.New()
	if err != nil {
		return err
	}

	n := naming.New(deb)

	err = steps.Create(dock, deb, n)
	if err != nil {
		return err
	}

	if start {
		err = steps.Start(dock, deb, n)
		if err != nil {
			return err
		}
	}

	return nil
}
