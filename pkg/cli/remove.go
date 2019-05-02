package cli

import (
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/stepping"
)

var stepRemove = &stepping.Step{
	Name: "remove",
	Run:  runRemove,
	Description: []string{
		"Removes container.",
		"Nothing more.",
	},
}

func runRemove() error {
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
