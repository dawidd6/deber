package cli

import (
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/omap"
	"os"
	"path/filepath"
)

const (
	stepBuild   = "build"
	stepCreate  = "create"
	stepStart   = "start"
	stepTarball = "tarball"
	stepEnable  = "enable"
	stepUpdate  = "update"
	stepDeps    = "deps"
	stepDisable = "disable"
	stepPackage = "package"
	stepTest    = "test"
	stepArchive = "archive"
	stepScan    = "scan"
	stepStop    = "stop"
	stepRemove  = "remove"

	stepCheckOptional = "check"
	stepShellOptional = "shell"
)

func getSteps() *omap.OrderedMap {
	steps := new(omap.OrderedMap)

	steps.Append(stepBuild, runBuild)
	steps.Append(stepCreate, runCreate)
	steps.Append(stepStart, runStart)
	steps.Append(stepTarball, runTarball)
	steps.Append(stepEnable, runEnable)
	steps.Append(stepUpdate, runUpdate)
	steps.Append(stepDeps, runDeps)
	steps.Append(stepDisable, runDisable)
	steps.Append(stepPackage, runPackage)
	steps.Append(stepTest, runTest)
	steps.Append(stepArchive, runArchive)
	steps.Append(stepScan, runScan)
	steps.Append(stepStop, runStop)
	steps.Append(stepRemove, runRemove)

	return steps
}

// runBuild function determines parent image name by querying DockerHub API
// for available "debian" and "ubuntu" tags and confronting them with
// debian/changelog's target distribution.
//
// At last it commands Docker Engine to build image.
func runBuild(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Building image")

	isImageBuilt, err := dock.IsImageBuilt(name.Image)
	if err != nil {
		return log.FailE(err)
	}
	if isImageBuilt {
		isImageOld, err := dock.IsImageOld(name.Image)
		if err != nil {
			return log.FailE(err)
		}
		if !isImageOld {
			return log.SkipE()
		}
	}

	for _, repo := range []string{"debian", "ubuntu"} {
		tags, err := docker.GetTags(repo)
		if err != nil {
			return log.FailE(err)
		}

		for _, tag := range tags {
			if tag.Name == deb.TargetDist {
				from := fmt.Sprintf("%s:%s", repo, deb.TargetDist)

				log.Drop()

				args := docker.ImageBuildArgs{
					From: from,
					Name: name.Image,
				}
				err := dock.ImageBuild(args)
				if err != nil {
					return log.FailE(err)
				}

				return log.DoneE()
			}
		}
	}

	return log.FailE(errors.New("distribution image not found"))
}

// runCreate function commands Docker Engine to create container.
func runCreate(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Creating container")

	isContainerCreated, err := dock.IsContainerCreated(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	if isContainerCreated {
		return log.SkipE()
	}

	args := docker.ContainerCreateArgs{
		SourceDir:  name.SourceDir,
		BuildDir:   name.BuildDir,
		ArchiveDir: name.ArchiveDir,
		CacheDir:   name.CacheDir,
		Image:      name.Image,
		Name:       name.Container,
		User:       fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
	}
	err = dock.ContainerCreate(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

// runStart function commands Docker Engine to start container.
func runStart(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Starting container")

	isContainerStarted, err := dock.IsContainerStarted(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	if isContainerStarted {
		return log.SkipE()
	}

	err = dock.ContainerStart(name.Container)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

// runTarball function moves orig upstream tarball from parent directory
// to build directory if package is not native.
func runTarball(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Moving tarball")

	tarball, err := deb.LocateTarball(name.SourceParentDir)
	if err != nil {
		return log.FailE(err)
	}

	if tarball == "" {
		return log.SkipE()
	}

	source := filepath.Join(name.SourceParentDir, tarball)
	target := filepath.Join(name.BuildDir, tarball)

	source, err = filepath.EvalSymlinks(source)
	if err != nil {
		return log.FailE(err)
	}

	err = os.Rename(source, target)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

// runEnable connects container to network.
func runEnable(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Enabling network")

	isContainerNetworkConnected, err := dock.IsContainerNetworkConnected(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	if isContainerNetworkConnected {
		return log.SkipE()
	}

	err = dock.ContainerEnableNetwork(name.Container)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

// runUpdate function executes apt-get update in container and
// creates "Packages" file in archive if not present.
func runUpdate(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Updating cache")

	log.Drop()

	file := filepath.Join(name.ArchiveDir, "Packages")
	info, _ := os.Stat(file)
	if info == nil {
		_, err := os.Create(file)
		if err != nil {
			return log.FailE(err)
		}
	}

	args := docker.ContainerExecArgs{
		Name:   name.Container,
		Cmd:    "apt-get update",
		AsRoot: true,
	}
	err := dock.ContainerExec(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

// runDeps function executes "mk-build-deps" in container.
func runDeps(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Installing dependencies")

	log.Drop()

	args := docker.ContainerExecArgs{
		Name:   name.Container,
		Cmd:    "mk-build-deps -ri",
		AsRoot: true,
	}
	err := dock.ContainerExec(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

// runDisable disconnects container from network.
func runDisable(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Disabling network")

	isContainerNetworkConnected, err := dock.IsContainerNetworkConnected(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	if !isContainerNetworkConnected {
		return log.SkipE()
	}

	err = dock.ContainerDisableNetwork(name.Container)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

// runPackage function first disables network in container,
// then executes "dpkg-buildpackage" and at the end,
// enables network back.
func runPackage(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Packaging software")

	log.Drop()

	args := docker.ContainerExecArgs{
		Name: name.Container,
		Cmd:  "dpkg-buildpackage" + " " + dpkgFlags,
	}
	err := dock.ContainerExec(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

// runTest function executes "debc", "debi" and "lintian" in container.
func runTest(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Testing package")

	log.Drop()

	commands := []string{
		"debc",
		"debi --with-depends",
		"lintian" + " " + lintianFlags,
	}

	for _, cmd := range commands {
		args := docker.ContainerExecArgs{
			Name:   name.Container,
			Cmd:    cmd,
			AsRoot: true,
		}
		err := dock.ContainerExec(args)
		if err != nil {
			return log.FailE(err)
		}
	}

	return log.DoneE()
}

// runArchive function moves successful build to archive by overwriting.
func runArchive(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Archiving build")

	info, _ := os.Stat(name.ArchivePackageDir)
	if info != nil {
		err := os.RemoveAll(name.ArchivePackageDir)
		if err != nil {
			return log.FailE(err)
		}
	}

	err := os.Rename(name.BuildDir, name.ArchivePackageDir)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

// runScan function executes "dpkg-scanpackages" in container and archive.
func runScan(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Scanning archive")

	log.Drop()

	args := docker.ContainerExecArgs{
		Name:    name.Container,
		Cmd:     "dpkg-scanpackages -m . > Packages",
		WorkDir: docker.ContainerArchiveDir,
	}
	err := dock.ContainerExec(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

// runStop function commands Docker Engine to stop container.
func runStop(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Stopping container")

	isContainerStopped, err := dock.IsContainerStopped(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	if isContainerStopped {
		return log.SkipE()
	}

	err = dock.ContainerStop(name.Container)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

// runRemove function commands Docker Engine to remove container.
func runRemove(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Removing container")

	isContainerCreated, err := dock.IsContainerCreated(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	if !isContainerCreated {
		return log.SkipE()
	}

	err = dock.ContainerRemove(name.Container)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

// runShellOptional function interactively executes bash shell in container.
func runShellOptional(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Launching shell")

	log.Drop()

	args := docker.ContainerExecArgs{
		Interactive: true,
		AsRoot:      true,
		Name:        name.Container,
	}
	err := dock.ContainerExec(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

// runCheck function evaluates if package has been already built and
// is in archive, if it is, then it exits with 0 code.
func runCheckOptional(deb *debian.Debian, dock *docker.Docker, name *naming.Naming) error {
	log.Info("Checking archive")

	info, _ := os.Stat(name.ArchivePackageDir)
	if info != nil {
		log.Skip()
		os.Exit(0)
	}

	return log.DoneE()
}
