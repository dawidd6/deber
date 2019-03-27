package app

import (
	"fmt"
	doc "github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/spf13/cobra"
	"os"
	"pault.ag/go/debian/changelog"
	"strings"
	"time"
)

func run(cmd *cobra.Command, args []string) error {
	log.Info("Parsing Debian changelog")
	debian, err := changelog.ParseFileOne("debian/changelog")
	if err != nil {
		return err
	}

	log.Info("Connecting with Docker")
	docker, err := doc.New()
	if err != nil {
		return err
	}

	tarball, err := getTarball(debian.Source, debian.Version.Version)
	if err != nil && !debian.Version.IsNative() {
		return err
	}

	name := naming.New(
		cmd.Use,
		from,
		debian.Source,
		debian.Version.String(),
		tarball,
	)

	steps := []func(*doc.Docker, *naming.Naming) error{
		runBuild,
		runCreate,
		runStart,
		runPackage,
		runTest,
		runStop,
		runRemove,
		runMove,
	}

	if clean {
		err := runStop(docker, name)
		if err != nil {
			return err
		}

		err = runRemove(docker, name)
		if err != nil {
			return err
		}

		return nil
	}

	for i := range steps {
		err := steps[i](docker, name)
		if err != nil {
			return err
		}
	}

	return nil
}

func runBuild(docker *doc.Docker, name *naming.Naming) error {
	log.Info("Building image")

	isImageBuilt, err := docker.IsImageBuilt(name.Image())
	if err != nil {
		return err
	}
	if isImageBuilt {
		return nil
	}

	err = docker.BuildImage(name.Image(), from)
	if err != nil {
		return err
	}

	return nil
}

func runCreate(docker *doc.Docker, name *naming.Naming) error {
	log.Info("Creating container")

	isContainerCreated, err := docker.IsContainerCreated(name.Container())
	if err != nil {
		return err
	}
	if isContainerCreated {
		return nil
	}

	err = docker.CreateContainer(name)
	if err != nil {
		return err
	}

	return nil
}

func runStart(docker *doc.Docker, name *naming.Naming) error {
	log.Info("Starting container")

	isContainerStarted, err := docker.IsContainerStarted(name.Container())
	if err != nil {
		return err
	}
	if isContainerStarted {
		return nil
	}

	err = docker.StartContainer(name.Container())
	if err != nil {
		return err
	}

	return nil
}

func runPackage(docker *doc.Docker, name *naming.Naming) error {
	log.Info("Packaging software")

	if name.Tarball() != "" {
		err := os.Rename(name.HostSourceSourceTarballFile(), name.HostBuildTargetTarballFile())
		if err != nil {
			return err
		}
	}

	file := fmt.Sprintf("%s/last-updated", name.HostBuildCacheDir())
	info, err := os.Stat(file)
	if info == nil || repo != "" || time.Now().Sub(info.ModTime()).Seconds() > update.Seconds() {
		err = docker.ExecContainer(name.Container(), "sudo", "apt-get", "update")
		if err != nil {
			return err
		}

		_, err := os.Create(file)
		if err != nil {
			return err
		}
	}

	err = docker.ExecContainer(name.Container(), "sudo", "mk-build-deps", "-ri", "-t", "apty")
	if err != nil {
		return err
	}

	flags := strings.Split(dpkgFlags, " ")
	command := append([]string{"dpkg-buildpackage"}, flags...)
	err = docker.ExecContainer(name.Container(), command...)
	if err != nil {
		return err
	}

	return nil
}

func runTest(docker *doc.Docker, name *naming.Naming) error {
	log.Info("Testing package")

	err := docker.ExecContainer(name.Container(), "debc")
	if err != nil {
		return err
	}

	err = docker.ExecContainer(name.Container(), "sudo", "debi", "--with-depends", "--tool", "apty")
	if err != nil {
		return err
	}

	flags := strings.Split(lintianFlags, " ")
	command := append([]string{"lintian"}, flags...)
	err = docker.ExecContainer(name.Container(), command...)
	if err != nil {
		return err
	}

	return nil
}

func runStop(docker *doc.Docker, name *naming.Naming) error {
	log.Info("Stopping container")

	isContainerStopped, err := docker.IsContainerStopped(name.Container())
	if err != nil {

		return err
	}
	if isContainerStopped {
		return nil
	}

	err = docker.StopContainer(name.Container())
	if err != nil {
		return err
	}

	return nil
}

func runRemove(docker *doc.Docker, name *naming.Naming) error {
	log.Info("Removing container")

	isContainerCreated, err := docker.IsContainerCreated(name.Container())
	if err != nil {
		return err
	}
	if !isContainerCreated {
		return nil
	}

	err = docker.RemoveContainer(name.Container())
	if err != nil {
		return err
	}

	return nil
}

func runMove(docker *doc.Docker, name *naming.Naming) error {
	log.Info("Moving output")

	info, err := os.Stat(name.HostArchiveFromOutputDir())
	if info != nil {
		return nil
	}

	err = os.Rename(name.HostBuildOutputDir(), name.HostArchiveFromOutputDir())
	if err != nil {
		return err
	}

	return nil
}
