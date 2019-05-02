package walking

import (
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
)

func ShellOptional(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
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
