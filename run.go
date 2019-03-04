package main

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

func run(cmd *cobra.Command, args []string) error {
	if stepping {
		for i := range steps {
			fmt.Println(steps[i].label)
		}
		return nil
	}

	if len(args) < 2 {
		return errors.New("need to specify OS and DIST")
	}

	if len(args) > 2 {
		dpkgOptions = args[2:]
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

	err = docker.BuildImage(names.Image(), names.os+":"+names.dist)
	if err != nil {
		LogFail()
		return err
	}

	LogDone()
	return nil
}

func runCreate() error {
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

func runStart() error {
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

func runPackage() error {
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

	command := []string{"dpkg-buildpackage", "-tc"}
	if len(dpkgOptions) > 0 {
		command = append(command, dpkgOptions...)
	}
	err = docker.ExecContainer(names.Container(), command...)
	if err != nil {
		LogFail()
		return err
	}

	LogDone()
	return nil
}

func runTest() error {
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

func runStop() error {
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

func runRemove() error {
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
