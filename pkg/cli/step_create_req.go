package cli

import (
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/stepping"
)

var stepCreate = &stepping.Step{
	Name: "create",
	Run:  runCreate,
	Description: []string{
		"Creates container and makes needed directories on host system.",
		"Will fail if image is nonexistent.",
	},
}

func runCreate() error {
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
	}
	err = dock.ContainerCreate(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}
