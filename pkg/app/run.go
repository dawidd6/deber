package app

import (
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/spf13/cobra"
	"strings"
	"syscall"
)

var (
	deb  *debian.Debian
	dock *docker.Docker
	name *naming.Naming
)

func parse(cmd *cobra.Command, args []string) error {
	if showSteps {
		for i := range steps {
			fmt.Printf("%s - %s\n", steps[i].label, steps[i].description)
		}
		syscall.Exit(0)
	}

	if withSteps != "" && withoutSteps != "" {
		return errors.New("can't specify with and without steps together")
	}

	if withSteps != "" {
		for i := range steps {
			if !strings.Contains(withSteps, steps[i].label) {
				steps[i].disabled = true
			}
		}
	}

	if withoutSteps != "" {
		for i := range steps {
			if strings.Contains(withoutSteps, steps[i].label) {
				steps[i].disabled = true
			}
		}
	}

	return nil
}

func run(cmd *cobra.Command, args []string) error {
	var err error

	err = parse(cmd, args)
	if err != nil {
		return err
	}

	logInfo("Parsing Debian changelog")
	deb, err = debian.New()
	if err != nil {
		return err
	}

	logInfo("Connecting with Docker")
	dock, err = docker.New()
	if err != nil {
		return err
	}

	name = naming.New(
		program,
		os,
		dist,
		deb.Source,
		deb.Version,
	)

	for i := range steps {
		if !steps[i].disabled {
			if err := steps[i].run(); err != nil {
				return err
			}
		}
	}

	return nil
}

func runBuild() error {
	logInfo("Building image")

	isImageBuilt, err := dock.IsImageBuilt(name.Image())
	if err != nil {
		return err
	}
	if isImageBuilt {
		return nil
	}

	err = dock.BuildImage(name.Image(), name.From())
	if err != nil {
		return err
	}

	return nil
}

func runCreate() error {
	logInfo("Creating container")

	isContainerCreated, err := dock.IsContainerCreated(name.Container())
	if err != nil {
		return err
	}
	if isContainerCreated {
		return nil
	}

	err = dock.CreateContainer(name.Container(), name.Image(), name.BuildDir(), repo, deb.Tarball)
	if err != nil {
		return err
	}

	return nil
}

func runStart() error {
	logInfo("Starting container")

	isContainerStarted, err := dock.IsContainerStarted(name.Container())
	if err != nil {
		return err
	}
	if isContainerStarted {
		return nil
	}

	err = dock.StartContainer(name.Container())
	if err != nil {
		return err
	}

	return nil
}

func runPackage() error {
	logInfo("Packaging software")

	err := dock.ExecContainer(name.Container(), "sudo", "apt-get", "update")
	if err != nil {
		return err
	}

	err = dock.ExecContainer(name.Container(), "sudo", "mk-build-deps", "-ri", "-t", "apty")
	if err != nil {
		return err
	}

	networks := []string{""}
	if !network {
		networks, err = dock.DisconnectAllNetworks(name.Container())
		if err != nil {
			return err
		}
	}

	flags := strings.Split(dpkgFlags, " ")
	command := append([]string{"dpkg-buildpackage"}, flags...)
	err = dock.ExecContainer(name.Container(), command...)
	if err != nil {
		return err
	}

	if !network {
		err = dock.ConnectNetworks(name.Container(), networks)
		if err != nil {
			return err
		}
	}

	return nil
}

func runTest() error {
	logInfo("Testing package")

	err := dock.ExecContainer(name.Container(), "sudo", "debi", "--with-depends", "--tool", "apty")
	if err != nil {
		return err
	}

	err = dock.ExecContainer(name.Container(), "debc")
	if err != nil {
		return err
	}

	flags := strings.Split(lintianFlags, " ")
	command := append([]string{"lintian"}, flags...)
	err = dock.ExecContainer(name.Container(), command...)
	if err != nil {
		return err
	}

	return nil
}

func runStop() error {
	logInfo("Stopping container")

	isContainerStopped, err := dock.IsContainerStopped(name.Container())
	if err != nil {
		return err
	}
	if isContainerStopped {
		return nil
	}

	err = dock.StopContainer(name.Container())
	if err != nil {
		return err
	}

	return nil
}

func runRemove() error {
	logInfo("Removing container")

	isContainerCreated, err := dock.IsContainerCreated(name.Container())
	if err != nil {
		return err
	}
	if !isContainerCreated {
		return nil
	}

	err = dock.RemoveContainer(name.Container())
	if err != nil {
		return err
	}

	return nil
}
