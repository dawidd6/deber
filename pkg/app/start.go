package app

import (
	"github.com/dawidd6/deber/pkg/logger"
)

func runStart(os, dist string) error {
	logger.Info("Starting container")

	isContainerStarted, err := dock.IsContainerStarted(names.Container())
	if err != nil {
		logger.Fail()
		return err
	}
	if isContainerStarted {
		logger.Skip()
		return nil
	}

	err = dock.StartContainer(names.Container())
	if err != nil {
		logger.Fail()
		return err
	}

	logger.Done()
	return nil
}
