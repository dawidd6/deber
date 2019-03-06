package app

import (
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/spf13/cobra"
	"strings"
)

var (
	deb  *debian.Debian
	dock *docker.Docker
	name *naming.Naming
)

func parse(args []string) error {
	if showStepsFlag {
		for i := range steps {
			fmt.Printf("%s - %s\n", steps[i].label, steps[i].description)
		}
		return nil
	}

	if len(args) < 2 {
		return errors.New("need to specify OS and DIST")
	}

	if len(args) > 2 {
		dpkgOptions = args[2:]
	}

	if withStepsFlag != "" && withoutStepsFlag != "" {
		return errors.New("can't specify with and without steps together")
	}

	if withStepsFlag != "" {
		for i := range steps {
			if !strings.Contains(withStepsFlag, steps[i].label) {
				steps[i].disabled = true
			}
		}
	}

	if withoutStepsFlag != "" {
		for i := range steps {
			if strings.Contains(withoutStepsFlag, steps[i].label) {
				steps[i].disabled = true
			}
		}
	}

	return nil
}

func run(cmd *cobra.Command, args []string) error {
	var err error

	err = parse(args)
	if err != nil {
		return err
	}

	logInfo("Parsing Debian changelog")
	deb, err = debian.New()
	if err != nil {
		logFail()
		return err
	}
	logDone()

	logInfo("Connecting with Docker")
	dock, err = docker.New(verboseFlag)
	if err != nil {
		logFail()
		return err
	}
	logDone()

	name = naming.New(
		program,
		args[0],
		args[1],
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
		logFail()
		return err
	}
	if isImageBuilt {
		logSkip()
		return nil
	}

	err = dock.BuildImage(name.Image(), name.From())
	if err != nil {
		logFail()
		return err
	}

	logDone()
	return nil
}

func runCreate() error {
	logInfo("Creating container")

	isContainerCreated, err := dock.IsContainerCreated(name.Container())
	if err != nil {
		logFail()
		return err
	}
	if isContainerCreated {
		logSkip()
		return nil
	}

	err = dock.CreateContainer(name.Container(), name.Image(), name.BuildDir(), deb.Tarball)
	if err != nil {
		logFail()
		return err
	}

	logDone()
	return nil
}

func runStart() error {
	logInfo("Starting container")

	isContainerStarted, err := dock.IsContainerStarted(name.Container())
	if err != nil {
		logFail()
		return err
	}
	if isContainerStarted {
		logSkip()
		return nil
	}

	err = dock.StartContainer(name.Container())
	if err != nil {
		logFail()
		return err
	}

	logDone()
	return nil
}

func runPackage() error {
	logInfo("Packaging software")

	err := dock.ExecContainer(name.Container(), "sudo", "apt-get", "update")
	if err != nil {
		logFail()
		return err
	}

	err = dock.ExecContainer(name.Container(), "sudo", "mk-build-deps", "-ri", "-t", "apt-get -y")
	if err != nil {
		logFail()
		return err
	}

	command := []string{"dpkg-buildpackage", "-tc"}
	if len(dpkgOptions) > 0 {
		command = append(command, dpkgOptions...)
	}
	err = dock.ExecContainer(name.Container(), command...)
	if err != nil {
		logFail()
		return err
	}

	logDone()
	return nil
}

func runTest() error {
	logInfo("Testing package")

	err := dock.ExecContainer(name.Container(), "sudo", "debi")
	if err != nil {
		logFail()
		return err
	}

	err = dock.ExecContainer(name.Container(), "debc")
	if err != nil {
		logFail()
		return err
	}

	err = dock.ExecContainer(name.Container(), "lintian")
	if err != nil {
		logFail()
		return err
	}

	logDone()
	return nil
}

func runStop() error {
	logInfo("Stopping container")

	isContainerStopped, err := dock.IsContainerStopped(name.Container())
	if err != nil {
		logFail()
		return err
	}
	if isContainerStopped {
		logSkip()
		return nil
	}

	err = dock.StopContainer(name.Container())
	if err != nil {
		logFail()
		return err
	}

	logDone()
	return nil
}

func runRemove() error {
	logInfo("Removing container")

	isContainerCreated, err := dock.IsContainerCreated(name.Container())
	if err != nil {
		logFail()
		return err
	}
	if !isContainerCreated {
		logSkip()
		return nil
	}

	err = dock.RemoveContainer(name.Container())
	if err != nil {
		logFail()
		return err
	}

	logDone()
	return nil
}
