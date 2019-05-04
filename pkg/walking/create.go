package walking

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
	"os"
)

var user = fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())

// StepCreate defines create step.
var StepCreate = &stepping.Step{
	Name: "create",
	Run:  Create,
	Description: []string{
		"Creates container and makes needed directories on host system.",
		"Will fail if image is nonexistent.",
	},
}

// Create function commands Docker Engine to create container.
func Create(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Creating container")

	isContainerCreated, err := dock.IsContainerCreated(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	if isContainerCreated {
		return log.SkipE()
	}

	args := docker.ContainerCreateArgs{
		SourceDir:  name.SourceDir,
		BuildDir:   name.BuildDir,
		ArchiveDir: name.ArchiveDir,
		CacheDir:   name.CacheDir,
		Image:      name.Image,
		Name:       name.Container,
		User:       user,
	}
	err = dock.ContainerCreate(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}
