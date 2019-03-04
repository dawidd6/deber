package app

import (
	"github.com/dawidd6/deber/pkg/logger"
)

func runBuild(os, dist string) error {
	logger.Info("Building image")

	isImageBuilt, err := dock.IsImageBuilt(names.Image())
	if err != nil {
		logger.Fail()
		return err
	}
	if isImageBuilt {
		logger.Skip()
		return nil
	}

	err = dock.BuildImage(names.Image(), os+":"+dist)
	if err != nil {
		logger.Fail()
		return err
	}

	logger.Done()
	return nil
}
