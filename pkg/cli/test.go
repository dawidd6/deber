package cli

import (
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/stepping"
)

var stepTest = &stepping.Step{
	Name: "test",
	Run:  runTest,
	Description: []string{
		"Runs series of commands in container:",
		"  - debc",
		"  - debi",
		"  - lintian",
		"Options passed to `lintian` are taken from environment variable",
		"Current `lintian` options: " + lintianFlags,
	},
}

func runTest() error {
	log.Info("Testing package")

	log.Drop()

	commands := []string{
		"debc",
		"sudo debi --with-depends",
		"lintian" + " " + lintianFlags,
	}

	for _, cmd := range commands {
		args := docker.ContainerExecArgs{
			Name: name.Container,
			Cmd:  cmd,
		}
		err := dock.ContainerExec(args)
		if err != nil {
			return log.FailE(err)
		}
	}

	return log.DoneE()
}