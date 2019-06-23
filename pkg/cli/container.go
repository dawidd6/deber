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
}

func runContainer(cmd *cobra.Command, args []string) error {
	flag := false

	if flagContainerList {
		flag = true

		containers, err := docker.ContainerList(app.Name)
		if err != nil {
			return err
		}

		for i := range containers {
			fmt.Println(containers[i])
		}
	}

	if flagContainerCreate {
		flag = true

		err := steps.Create()
		if err != nil {
			return err
		}
	}

	if flagContainerStart {
		flag = true

		err := steps.Start()
		if err != nil {
			return err
		}
	}

	if flagContainerShell {
		flag = true

		err := steps.ShellOptional()
		if err != nil {
			return err
		}
	}

	if flagContainerStop {
		flag = true

		err := steps.Stop()
		if err != nil {
			return err
		}
	}

	if flagContainerRemove {
		flag = true

		err := steps.Remove()
		if err != nil {
			return err
		}
	}

	if flag {
		return nil
	}

	return cmd.Help()
}
