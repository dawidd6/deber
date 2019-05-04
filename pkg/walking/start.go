package walking

import (
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
)

// StepStart defines start step.
var StepStart = &stepping.Step{
	Name: "start",
	Run:  Start,
	Description: []string{
		"Starts previously created container.",
		"The entry command is `sleep inf`, which means that container",
		"will just sit there, doing nothing and waiting for commands.",
	},
}

// Start function commands Docker Engine to start container.
func Start(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Starting container")

	isContainerStarted, err := dock.IsContainerStarted(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	if isContainerStarted {
		return log.SkipE()
	}

	err = dock.ContainerStart(name.Container)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}
