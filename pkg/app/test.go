package app

import (
	"github.com/dawidd6/deber/pkg/logger"
)

func runTest(os, dist string) error {
	logger.Info("Testing package")

	err := dock.ExecContainer(names.Container(), "sudo", "debi")
	if err != nil {
		logger.Fail()
		return err
	}

	err = dock.ExecContainer(names.Container(), "debc")
	if err != nil {
		logger.Fail()
		return err
	}

	err = dock.ExecContainer(names.Container(), "lintian")
	if err != nil {
		logger.Fail()
		return err
	}

	logger.Done()
	return nil
}
