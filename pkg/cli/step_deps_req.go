package cli

import (
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/stepping"
)

var stepDeps = &stepping.Step{
	Name: "deps",
	Run:  runDeps,
	Description: []string{
		"Installs package's build dependencies in container.",
		"Runs `mk-build-deps` with appropriate options.",
	},
}

func runDeps() error {
	log.Info("Installing dependencies")

	log.Drop()

	args := docker.ContainerExecArgs{
		Name: name.Container,
		Cmd:  "sudo mk-build-deps -ri",
	}
	err := dock.ContainerExec(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}
