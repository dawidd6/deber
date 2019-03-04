package main

import (
	"errors"
	"github.com/spf13/cobra"
	"strings"
)

func pre(cmd *cobra.Command, args []string) error {
	var err error

	LogInfo("Parsing Debian changelog")
	debian, err = NewDebian()
	if err != nil {
		LogFail()
		return err
	}
	LogDone()

	LogInfo("Connecting with Docker")
	docker, err = NewDocker(verbose)
	if err != nil {
		LogFail()
		return err
	}
	LogDone()

	names = NewNaming(args[0], args[1], debian)

	return nil
}

func run(cmd *cobra.Command, args []string) error {
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

	for i := range steps {
		if !steps[i].disabled {
			err := steps[i].run(args[0], args[1])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func runBuild(os, dist string) error {
	LogInfo("Building image")

	isImageBuilt, err := docker.IsImageBuilt(names.Image())
	if err != nil {
		LogFail()
		return err
	}
	if isImageBuilt {
		LogSkip()
		return nil
	}

	err = docker.BuildImage(names.Image(), os+":"+dist)
	if err != nil {
		LogFail()
		return err
	}

	LogDone()
	return nil
}

func runCreate(os, dist string) error {
	LogInfo("Creating container")

	isContainerCreated, err := docker.IsContainerCreated(names.Container())
	if err != nil {
		LogFail()
		return err
	}
	if isContainerCreated {
		LogSkip()
		return nil
	}

	err = docker.CreateContainer(names.Container(), names.Image(), names.BuildDir(), debian.Tarball)
	if err != nil {
		LogFail()
		return err
	}

	LogDone()
	return nil
}

func runStart(os, dist string) error {
	LogInfo("Starting container")

	isContainerStarted, err := docker.IsContainerStarted(names.Container())
	if err != nil {
		LogFail()
		return err
	}
	if isContainerStarted {
		LogSkip()
		return nil
	}

	err = docker.StartContainer(names.Container())
	if err != nil {
		LogFail()
		return err
	}

	LogDone()
	return nil
}

func runPackage(os, dist string) error {
	LogInfo("Packaging software")

	err := docker.ExecContainer(names.Container(), "sudo", "apt-get", "update")
	if err != nil {
		LogFail()
		return err
	}

	err = docker.ExecContainer(names.Container(), "sudo", "mk-build-deps", "-ri", "-t", "apt-get -y")
	if err != nil {
		LogFail()
		return err
	}

	err = docker.ExecContainer(names.Container(), "dpkg-buildpackage", "-tc")
	if err != nil {
		LogFail()
		return err
	}

	LogDone()
	return nil
}

func runTest(os, dist string) error {
	LogInfo("Testing package")

	err := docker.ExecContainer(names.Container(), "sudo", "debi")
	if err != nil {
		LogFail()
		return err
	}

	err = docker.ExecContainer(names.Container(), "debc")
	if err != nil {
		LogFail()
		return err
	}

	err = docker.ExecContainer(names.Container(), "lintian")
	if err != nil {
		LogFail()
		return err
	}

	LogDone()
	return nil
}

func runStop(os, dist string) error {
	LogInfo("Stopping container")

	isContainerStopped, err := docker.IsContainerStopped(names.Container())
	if err != nil {
		LogFail()
		return err
	}
	if isContainerStopped {
		LogSkip()
		return nil
	}

	err = docker.StopContainer(names.Container())
	if err != nil {
		LogFail()
		return err
	}

	LogDone()
	return nil
}

func runRemove(os, dist string) error {
	LogInfo("Removing container")

	isContainerCreated, err := docker.IsContainerCreated(names.Container())
	if err != nil {
		LogFail()
		return err
	}
	if !isContainerCreated {
		LogSkip()
		return nil
	}

	err = docker.RemoveContainer(names.Container())
	if err != nil {
		LogFail()
		return err
	}

	LogDone()
	return nil
}
