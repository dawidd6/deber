package app

import (
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

var (
	deb  *debian.Debian
	dock *docker.Docker

	deContainer string
	deImage     string
)

func parse(cmd *cobra.Command, args []string) error {
	if showSteps {
		for i := range steps {
			fmt.Printf("%s - %s\n", steps[i].label, steps[i].description)
		}
		os.Exit(0)
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

	deContainer = naming.Container(program, image, deb.Source, deb.Version)
	deImage = naming.Image(program, image)

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

	isImageBuilt, err := dock.IsImageBuilt(deImage)
	if err != nil {
		return err
	}
	if isImageBuilt {
		return nil
	}

	err = dock.BuildImage(deImage, image)
	if err != nil {
		return err
	}

	return nil
}

func runCreate() error {
	logInfo("Creating container")

	isContainerCreated, err := dock.IsContainerCreated(deContainer)
	if err != nil {
		return err
	}
	if isContainerCreated {
		return nil
	}

	err = dock.CreateContainer(deContainer, deImage, repo, deb.Tarball)
	if err != nil {
		return err
	}

	return nil
}

func runStart() error {
	logInfo("Starting container")

	isContainerStarted, err := dock.IsContainerStarted(deContainer)
	if err != nil {
		return err
	}
	if isContainerStarted {
		return nil
	}

	err = dock.StartContainer(deContainer)
	if err != nil {
		return err
	}

	return nil
}

func runPackage() error {
	logInfo("Packaging software")

	file := fmt.Sprintf("%s/last-updated", docker.HostCacheDir(deImage))
	info, err := os.Stat(file)
	if info == nil || time.Now().Sub(info.ModTime()).Seconds() > update.Seconds() {
		err = dock.ExecContainer(deContainer, "sudo", "apt-get", "update")
		if err != nil {
			return err
		}

		_, err := os.Create(file)
		if err != nil {
			return err
		}
	}

	err = dock.ExecContainer(deContainer, "sudo", "mk-build-deps", "-ri", "-t", "apty")
	if err != nil {
		return err
	}

	flags := strings.Split(dpkgFlags, " ")
	command := append([]string{"dpkg-buildpackage"}, flags...)
	err = dock.ExecContainer(deContainer, command...)
	if err != nil {
		return err
	}

	return nil
}

func runTest() error {
	logInfo("Testing package")

	err := dock.ExecContainer(deContainer, "debc")
	if err != nil {
		return err
	}

	err = dock.ExecContainer(deContainer, "sudo", "debi", "--with-depends", "--tool", "apty")
	if err != nil {
		return err
	}

	flags := strings.Split(lintianFlags, " ")
	command := append([]string{"lintian"}, flags...)
	err = dock.ExecContainer(deContainer, command...)
	if err != nil {
		return err
	}

	return nil
}

func runStop() error {
	logInfo("Stopping container")

	isContainerStopped, err := dock.IsContainerStopped(deContainer)
	if err != nil {
		return err
	}
	if isContainerStopped {
		return nil
	}

	err = dock.StopContainer(deContainer)
	if err != nil {
		return err
	}

	return nil
}

func runRemove() error {
	logInfo("Removing container")

	isContainerCreated, err := dock.IsContainerCreated(deContainer)
	if err != nil {
		return err
	}
	if !isContainerCreated {
		return nil
	}

	err = dock.RemoveContainer(deContainer)
	if err != nil {
		return err
	}

	return nil
}
