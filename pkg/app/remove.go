package app

import (
	"github.com/dawidd6/deber/pkg/logger"
)

func runRemove(os, dist string) error {
	logger.Info("Removing container")

	isContainerCreated, err := dock.IsContainerCreated(names.Container())
	if err != nil {
		logger.Fail()
		return err
	}
	if !isContainerCreated {
		logger.Skip()
		return nil
	}

	err = dock.RemoveContainer(names.Container())
	if err != nil {
		logger.Fail()
		return err
	}

	logger.Done()
	return nil
}
