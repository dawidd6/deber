package commands

import (
	"errors"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"os"
)

var cmdBuild = &cobra.Command{
	Use:               "build OS DIST",
	Short:             "build Docker image",
	Args:              cobra.ExactArgs(2),
	PersistentPreRunE: pre,
	RunE:              runBuild,
}

func runBuild(cmd *cobra.Command, args []string) error {
	logger.Info("Building image")

	writer := ioutil.Discard
	if verbose {
		writer = os.Stdout
	}

	isImageBuilt, err := dock.IsImageBuilt(names.Image())
	if err != nil {
		logger.Fail()
		return err
	}
	if isImageBuilt {
		logger.Skip()
		return nil
	}

	dockerfile, err := docker.GetDockerfile(args[0], args[1])
	if err != nil {
		logger.Fail()
		return err
	}

	response, err := dock.BuildImage(names.Image(), dockerfile)
	if err != nil {
		logger.Fail()
		return err
	}

	_, err = io.Copy(writer, response.Body)
	if err != nil {
		logger.Fail()
		return err
	}

	err = response.Body.Close()
	if err != nil {
		logger.Fail()
		return err
	}

	isImageBuilt, err = dock.IsImageBuilt(names.Image())
	if err != nil {
		logger.Fail()
		return err
	}

	if !isImageBuilt {
		logger.Fail()
		return errors.New("image didn't built successfully")
	}

	logger.Done()
	return nil
}
