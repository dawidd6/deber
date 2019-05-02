package walking

import (
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/stepping"
)

var StepRemove = &stepping.Step{
	Name: "remove",
	Run:  Remove,
	Description: []string{
		"Removes container.",
		"Nothing more.",
	},
}

func Remove(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Removing container")

	isContainerCreated, err := dock.IsContainerCreated(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	if !isContainerCreated {
		return log.SkipE()
	}

	err = dock.ContainerRemove(name.Container)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}
