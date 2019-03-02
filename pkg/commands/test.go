package commands

import (
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"os"
)

var cmdTest = &cobra.Command{
	Use:               "test OS DIST",
	Short:             "test package in Docker container",
	Args:              cobra.ExactArgs(2),
	PersistentPreRunE: pre,
	RunE:              runTest,
}

func runTest(cmd *cobra.Command, args []string) error {
	logger.Info("Testing package")

	writer := ioutil.Discard
	if verbose {
		writer = os.Stdout
	}

	hijack, err := dock.ExecContainer(names.Container(), "sudo", "debi")
	if err != nil {
		logger.Fail()
		return err
	}

	_, err = io.Copy(writer, hijack.Reader)
	if err != nil {
		logger.Fail()
		return err
	}
	defer hijack.Close()

	hijack, err = dock.ExecContainer(names.Container(), "debc")
	if err != nil {
		logger.Fail()
		return err
	}

	_, err = io.Copy(writer, hijack.Reader)
	if err != nil {
		logger.Fail()
		return err
	}
	defer hijack.Close()

	hijack, err = dock.ExecContainer(names.Container(), "lintian")
	if err != nil {
		logger.Fail()
		return err
	}

	_, err = io.Copy(writer, hijack.Reader)
	if err != nil {
		logger.Fail()
		return err
	}
	defer hijack.Close()

	logger.Done()
	return nil
}
