package commands

import (
	"deber/pkg/logger"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"os"
)

var cmdPackage = &cobra.Command{
	Use:               "package OS DIST",
	Short:             "package Docker container",
	Args:              cobra.ExactArgs(2),
	PersistentPreRunE: pre,
	RunE:              runPackage,
}

func runPackage(cmd *cobra.Command, args []string) error {
	logger.Info("Packaging software")

	writer := ioutil.Discard
	if verbose {
		writer = os.Stdout
	}

	hijack, err := dock.ExecContainer(names.Container(), "sudo", "apt-get", "update")
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

	hijack, err = dock.ExecContainer(names.Container(), "sudo", "mk-build-deps", "-ri", "-t", "apt-get -y")
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

	hijack, err = dock.ExecContainer(names.Container(), "dpkg-buildpackage", "-tc")
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
