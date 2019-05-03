package walking

import (
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
	"os"
)

var lintianFlags = os.Getenv("DEBER_LINTIAN_FLAGS")

// StepTest defines test step
var StepTest = &stepping.Step{
	Name: "test",
	Run:  Test,
	Description: []string{
		"Runs series of commands in container:",
		"  - debc",
		"  - debi",
		"  - lintian",
		"Options passed to `lintian` are taken from environment variable",
		"Current `lintian` options: " + lintianFlags,
	},
}

// Test function executes "debc", "debi" and "lintian" in container
func Test(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Testing package")

	log.Drop()

	commands := []string{
		"debc",
		"debi --with-depends",
		"lintian" + " " + lintianFlags,
	}

	for _, cmd := range commands {
		args := docker.ContainerExecArgs{
			Name:   name.Container,
			Cmd:    cmd,
			AsRoot: true,
		}
		err := dock.ContainerExec(args)
		if err != nil {
			return log.FailE(err)
		}
	}

	return log.DoneE()
}
