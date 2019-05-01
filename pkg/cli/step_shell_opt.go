package cli

import "github.com/dawidd6/deber/pkg/docker"

func runShellOptional() error {
	args := docker.ContainerExecArgs{
		Interactive: true,
		Name:        name.Container,
	}
	err := dock.ContainerExec(args)
	if err != nil {
		return err
	}

	return nil
}
