package cli

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
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
	cmdRoot.AddCommand(
		cmdContainer,
	)

	cmdContainer.Flags().BoolVarP(&flagContainerCreate, "create", "c", flagContainerCreate, "")
	cmdContainer.Flags().BoolVarP(&flagContainerStart, "start", "t", flagContainerStart, "")
	cmdContainer.Flags().BoolVarP(&flagContainerRemove, "remove", "r", flagContainerRemove, "")
	cmdContainer.Flags().BoolVarP(&flagContainerStop, "stop", "p", flagContainerStop, "")
	cmdContainer.Flags().BoolVarP(&flagContainerList, "list", "l", flagContainerList, "")
	cmdContainer.Flags().BoolVarP(&flagContainerCreate, "shell", "s", flagContainerShell, "")
}

func runContainer(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

func runContainerCreate(cmd *cobra.Command, args []string) error {
	err = steps.Create(dock, deb, n)
	if err != nil {
		return err
	}

	if flagStartContainer {
		err = steps.Start(dock, deb, n)
		if err != nil {
			return err
		}
	}

	return nil
}

func runContainerRemove(cmd *cobra.Command, args []string) error {
	if flagStopContainer {
		err := steps.Stop(dock, deb, n)
		if err != nil {
			return err
		}
	}

	err := steps.Remove(dock, deb, n)
	if err != nil {
		return err
	}

	return nil
}

func runContainerShell(cmd *cobra.Command, args []string) error {
	return steps.ShellOptional(dock, deb, n)
}

func runContainerList(cmd *cobra.Command, args []string) error {
	containers, err := dock.ContainerList(app.Name)
	if err != nil {
		return err
	}

	for i := range containers {
		fmt.Println(containers[i])
	}

	return nil
}
