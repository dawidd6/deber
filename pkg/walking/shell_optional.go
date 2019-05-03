package walking

import (
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
)

// StepShellOptional defines optional shell step
var StepShellOptional = &stepping.Step{
	Name:     "shell",
	Run:      ShellOptional,
	Optional: true,
	Excluded: true,
	Description: []string{
		"Runs interactive bash shell in container.",
		"This step is optional and not executed by default.",
	},
}

// ShellOptional function interactively executes bash shell in container
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
