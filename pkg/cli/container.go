package cli

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/spf13/cobra"
)

var (
	flagContainerCreate bool
	flagContainerStart  bool
	flagContainerRemove bool
	flagContainerStop   bool
	flagContainerList   bool
	flagContainerShell  bool
	flagContainerPrune  bool
)

var (
	cmdContainer = &cobra.Command{
		Use:   "container",
		Short: "",
		RunE:  runContainer,
	}
)

func init() {
	cmdContainer.Flags().BoolVar(&flagContainerCreate, "create", flagContainerCreate, "")
	cmdContainer.Flags().BoolVar(&flagContainerStart, "start", flagContainerStart, "")
	cmdContainer.Flags().BoolVar(&flagContainerRemove, "remove", flagContainerRemove, "")
	cmdContainer.Flags().BoolVar(&flagContainerStop, "stop", flagContainerStop, "")
	cmdContainer.Flags().BoolVar(&flagContainerList, "list", flagContainerList, "")
	cmdContainer.Flags().BoolVar(&flagContainerShell, "shell", flagContainerShell, "")
	cmdContainer.Flags().BoolVar(&flagContainerPrune, "prune", flagContainerPrune, "")
}

func runContainer(cmd *cobra.Command, args []string) error {
	if flagContainerPrune {
		containers, err := docker.ContainerList(app.Name)
		if err != nil {
			return err
		}

		for i := range containers {
			err = docker.ContainerStop(containers[i])
			if err != nil {
				return err
			}

			err = docker.ContainerRemove(containers[i])
			if err != nil {
				return err
			}
		}
	}

	if flagContainerList {
		containers, err := docker.ContainerList(app.Name)
		if err != nil {
			return err
		}

		for i := range containers {
			fmt.Println(containers[i])
		}
	}

	if flagContainerCreate {
		err := steps.Create()
		if err != nil {
			return err
		}
	}

	if flagContainerStart {
		err := steps.Start()
		if err != nil {
			return err
		}
	}

	if flagContainerShell {
		err := steps.ShellOptional()
		if err != nil {
			return err
		}
	}

	if flagContainerStop {
		err := steps.Stop()
		if err != nil {
			return err
		}
	}

	if flagContainerRemove {
		err := steps.Remove()
		if err != nil {
			return err
		}
	}

	if cmd.Flags().NFlag() > 0 {
		return nil
	}

	return cmd.Help()
}
