package cli

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
	"os"
)

var stepPackage = &stepping.Step{
	Name: "package",
	Run:  runPackage,
	Description: []string{
		"Runs `dpkg-buildpackage` in container.",
		"Options passed to `dpkg-buildpackage` are taken from environment variable",
		"Current `dpkg-buildpackage` options: " + dpkgFlags,
	},
}

func runPackage(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Packaging software")

	file := fmt.Sprintf("%s/%s", name.ArchiveDir, "Packages")
	info, _ := os.Stat(file)
	if info == nil {
		_, err := os.Create(file)
		if err != nil {
			return log.FailE(err)
		}
	}

	err := dock.ContainerDisableNetwork(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	defer dock.ContainerEnableNetwork(name.Container)

	log.Drop()

	args := docker.ContainerExecArgs{
		Name: name.Container,
		Cmd:  "dpkg-buildpackage" + " " + dpkgFlags,
	}
	err = dock.ContainerExec(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}
