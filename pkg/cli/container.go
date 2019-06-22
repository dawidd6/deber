package cli

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
	"github.com/dawidd6/deber/pkg/steps"
	"github.com/spf13/cobra"
)

var (
	cmdContainer = &cobra.Command{
		Use:   "container",
		Short: "",
		RunE:  runContainer,
	}

	cmdContainerCreate = &cobra.Command{
		Use:   "create",
		Short: "",
		RunE:  runContainerCreate,
	}

	cmdContainerRemove = &cobra.Command{
		Use:   "remove",
		Short: "",
		RunE:  runContainerRemove,
	}

	cmdContainerShell = &cobra.Command{
		Use:   "shell",
		Short: "",
		RunE:  runContainerShell,
	}

	cmdContainerList = &cobra.Command{
		Use:   "list",
		Short: "",
		RunE:  runContainerList,
	}
)

func init() {
	cmdRoot.AddCommand(
		cmdContainer,
	)
	cmdContainer.AddCommand(
		cmdContainerCreate,
		cmdContainerRemove,
		cmdContainerShell,
		cmdContainerList,
	)

	cmdContainerCreate.Flags().BoolVarP(&flagStartContainer, "start", "s", flagStartContainer, "")
	cmdContainerCreate.Flags().StringArrayVar(&steps.ExtraPackages, "extra-package", steps.ExtraPackages, "")

	cmdContainerRemove.Flags().BoolVarP(&flagStopContainer, "stop", "s", flagStopContainer, "")
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
