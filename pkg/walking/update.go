package walking

import (
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
	"os"
	"path/filepath"
)

// StepUpdate defines update step.
var StepUpdate = &stepping.Step{
	Name: "update",
	Run:  Update,
	Description: []string{
		"Updates apt's cache.",
		"Also creates empty `Packages` file in archive if nonexistent",
	},
}

// Update function executes apt-get update in container and
// creates "Packages" file in archive if not present.
func Update(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Updating cache")

	log.Drop()

	file := filepath.Join(name.ArchiveDir, "Packages")
	info, _ := os.Stat(file)
	if info == nil {
		_, err := os.Create(file)
		if err != nil {
			return log.FailE(err)
		}
	}

	args := docker.ContainerExecArgs{
		Name:   name.Container,
		Cmd:    "apt-get update",
		AsRoot: true,
	}
	err := dock.ContainerExec(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}
