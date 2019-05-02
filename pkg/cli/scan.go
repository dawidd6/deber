package cli

import (
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/stepping"
)

var stepScan = &stepping.Step{
	Name: "scan",
	Run:  runScan,
	Description: []string{
		"Scans available packages in archive and writes result to `Packages` file.",
		"This `Packages` file is then used by apt in container.",
	},
}

func runScan() error {
	log.Info("Scanning archive")

	log.Drop()

	args := docker.ContainerExecArgs{
		Name:    name.Container,
		Cmd:     "dpkg-scanpackages -m . > Packages",
		WorkDir: docker.ContainerArchiveDir,
	}
	err := dock.ContainerExec(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}
