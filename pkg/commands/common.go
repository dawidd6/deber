package commands

import (
	"deber/pkg/debian"
	"deber/pkg/docker"
	"deber/pkg/logger"
	"deber/pkg/naming"
	"github.com/spf13/cobra"
)

var (
	deb   *debian.Debian
	dock  *docker.Docker
	names *naming.Naming
)

func pre(cmd *cobra.Command, args []string) error {
	var err error

	logger.Info("Parsing Debian changelog")
	deb, err = debian.New()
	if err != nil {
		return err
	}
	logger.Done()

	names = naming.New(args[0], args[1], deb)

	logger.Info("Connecting with Docker")
	dock, err = docker.New()
	if err != nil {
		return err
	}
	logger.Done()

	return nil
}
