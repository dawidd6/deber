package app

import (
	"fmt"
	deb "github.com/dawidd6/deber/pkg/debian"
	doc "github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

var (
	container string
	image     string
)

func run(cmd *cobra.Command, args []string) error {
	var err error

	logInfo("Parsing Debian changelog")
	debian, err := deb.New()
	if err != nil {
		return err
	}

	logInfo("Connecting with Docker")
	docker, err := doc.New()
	if err != nil {
		return err
	}

	container = naming.Container(program, from, debian.Source, debian.Version)
	image = naming.Image(program, from)

	steps := []func(*doc.Docker, *deb.Debian) error{
		runBuild,
		runCreate,
		runStart,
		runPackage,
		runTest,
		runStop,
		runRemove,
	}

	for i := range steps {
		err := steps[i](docker, debian)
		if err != nil {
			return err
		}
	}

	return nil
}

func runBuild(docker *doc.Docker, debian *deb.Debian) error {
	logInfo("Building image")

	isImageBuilt, err := docker.IsImageBuilt(image)
	if err != nil {
		return err
	}
	if isImageBuilt {
		return nil
	}

	err = docker.BuildImage(image, from)
	if err != nil {
		return err
	}

	return nil
}

func runCreate(docker *doc.Docker, debian *deb.Debian) error {
	logInfo("Creating container")

	isContainerCreated, err := docker.IsContainerCreated(container)
	if err != nil {
		return err
	}
	if isContainerCreated {
		return nil
	}

	err = docker.CreateContainer(container, image, repo, debian.Tarball)
	if err != nil {
		return err
	}

	return nil
}

func runStart(docker *doc.Docker, debian *deb.Debian) error {
	logInfo("Starting container")

	isContainerStarted, err := docker.IsContainerStarted(container)
	if err != nil {
		return err
	}
	if isContainerStarted {
		return nil
	}

	err = docker.StartContainer(container)
	if err != nil {
		return err
	}

	return nil
}

func runPackage(docker *doc.Docker, debian *deb.Debian) error {
	logInfo("Packaging software")

	file := fmt.Sprintf("%s/last-updated", naming.HostCacheDir(image))
	info, err := os.Stat(file)
	if info == nil || time.Now().Sub(info.ModTime()).Seconds() > update.Seconds() {
		err = docker.ExecContainer(container, "sudo", "apt-get", "update")
		if err != nil {
			return err
		}

		_, err := os.Create(file)
		if err != nil {
			return err
		}
	}

	err = docker.ExecContainer(container, "sudo", "mk-build-deps", "-ri", "-t", "apty")
	if err != nil {
		return err
	}

	flags := strings.Split(dpkgFlags, " ")
	command := append([]string{"dpkg-buildpackage"}, flags...)
	err = docker.ExecContainer(container, command...)
	if err != nil {
		return err
	}

	return nil
}

func runTest(docker *doc.Docker, debian *deb.Debian) error {
	logInfo("Testing package")

	err := docker.ExecContainer(container, "debc")
	if err != nil {
		return err
	}

	err = docker.ExecContainer(container, "sudo", "debi", "--with-depends", "--tool", "apty")
	if err != nil {
		return err
	}

	flags := strings.Split(lintianFlags, " ")
	command := append([]string{"lintian"}, flags...)
	err = docker.ExecContainer(container, command...)
	if err != nil {
		return err
	}

	return nil
}

func runStop(docker *doc.Docker, debian *deb.Debian) error {
	logInfo("Stopping container")

	isContainerStopped, err := docker.IsContainerStopped(container)
	if err != nil {
		return err
	}
	if isContainerStopped {
		return nil
	}

	err = docker.StopContainer(container)
	if err != nil {
		return err
	}

	return nil
}

func runRemove(docker *doc.Docker, debian *deb.Debian) error {
	logInfo("Removing container")

	isContainerCreated, err := docker.IsContainerCreated(container)
	if err != nil {
		return err
	}
	if !isContainerCreated {
		return nil
	}

	err = docker.RemoveContainer(container)
	if err != nil {
		return err
	}

	return nil
}
