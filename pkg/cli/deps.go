package cli

import (
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
)

var StepDeps = &stepping.Step{
	Name: "deps",
	Run:  Deps,
	Description: []string{
		"Installs package's build dependencies in container.",
		"Runs `mk-build-deps` with appropriate options.",
	},
}

func Deps(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Installing dependencies")

	log.Drop()

	args := docker.ContainerExecArgs{
		Name:   name.Container,
		Cmd:    "mk-build-deps -ri",
		AsRoot: true,
	}
	err := dock.ContainerExec(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}
