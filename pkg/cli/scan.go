package cli

import (
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
)

var StepScan = &stepping.Step{
	Name: "scan",
	Run:  Scan,
	Description: []string{
		"Scans available packages in archive and writes result to `Packages` file.",
		"This `Packages` file is then used by apt in container.",
	},
}

func Scan(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
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
