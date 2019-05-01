package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dawidd6/deber/pkg/stepping"

	"github.com/dawidd6/deber/pkg/logger"

	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/spf13/cobra"
)

var (
	deb  *debian.Debian
	dock *docker.Docker
	name *naming.Naming
	log  *logger.Logger
)

func run(cmd *cobra.Command, args []string) error {
	var err error

	// SECTION: declare steps
	steps := stepping.Steps{
		{
			Name: "check",
			Run:  runCheck,
		}, {
			Name: "build",
			Run:  runBuild,
		}, {
			Name: "create",
			Run:  runCreate,
		}, {
			Name: "start",
			Run:  runStart,
		}, {
			Name: "tarball",
			Run:  runTarball,
		}, {
			Name: "update",
			Run:  runUpdate,
		}, {
			Name: "deps",
			Run:  runDeps,
		}, {
			Name: "package",
			Run:  runPackage,
		}, {
			Name: "test",
			Run:  runTest,
		}, {
			Name: "archive",
			Run:  runArchive,
		}, {
			Name: "scan",
			Run:  runScan,
		}, {
			Name: "stop",
			Run:  runStop,
		}, {
			Name: "remove",
			Run:  runRemove,
		},
	}

	// SECTION: include-exclude
	if include != "" && exclude != "" {
		return errors.New("can't specify --include and --exclude together")
	}
	if include != "" {
		err := steps.Include(strings.Split(include, ",")...)
		if err != nil {
			return err
		}
	}
	if exclude != "" {
		err := steps.Exclude(strings.Split(exclude, ",")...)
		if err != nil {
			return err
		}
	}

	// SECTION: init stuff
	log = logger.New(cmd.Use)

	deb, err = debian.ParseChangelog()
	if err != nil {
		return err
	}

	dock, err = docker.New()
	if err != nil {
		return err
	}

	name = naming.New(
		cmd.Use,
		deb.TargetDist,
		deb.SourceName,
		deb.PackageVersion,
		archiveDir,
	)

	// SECTION: handle bool options
	if shell {
		return runShellOptional()
	}

	// SECTION: run steps
	err = steps.Run()
	if err != nil {
		return err
	}

	return nil
}

func runShellOptional() error {
	err := runCreate()
	if err != nil {
		return err
	}

	err = runStart()
	if err != nil {
		return err
	}

	args := docker.ContainerExecArgs{
		Interactive: true,
		Name:        name.Container,
	}
	err = dock.ContainerExec(args)
	if err != nil {
		return err
	}

	return nil
}

func runCheck() error {
	log.Info("Checking archive")

	info, _ := os.Stat(name.ArchivePackageDir)
	if info != nil {
		log.Skip()
		os.Exit(0)
	}

	return log.DoneE()
}

func runBuild() error {
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

				args := docker.BuildImageArgs{
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

	return log.FailE(errors.New("dist image not found"))
}

func runCreate() error {
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
	}
	err = dock.ContainerCreate(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runStart() error {
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

func runTarball() error {
	log.Info("Moving tarball")

	tarball, err := deb.LocateTarball()
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

func runUpdate() error {
	log.Info("Updating cache")

	log.Drop()

	file := filepath.Join(name.ArchiveDir, "Packages")
	info, err := os.Stat(file)
	if info == nil {
		_, err := os.Create(file)
		if err != nil {
			return log.FailE(err)
		}
	}

	args := docker.ContainerExecArgs{
		Name: name.Container,
		Cmd:  "sudo apt-get update",
	}
	err = dock.ContainerExec(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runDeps() error {
	log.Info("Installing dependencies")

	log.Drop()

	args := docker.ContainerExecArgs{
		Name: name.Container,
		Cmd:  "sudo mk-build-deps -ri",
	}
	err := dock.ContainerExec(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runPackage() error {
	log.Info("Packaging software")

	file := fmt.Sprintf("%s/%s", name.ArchiveDir, "Packages")
	info, err := os.Stat(file)
	if info == nil {
		_, err := os.Create(file)
		if err != nil {
			return log.FailE(err)
		}
	}

	err = dock.ContainerDisableNetwork(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	defer dock.ContainerEnableNetwork(name.Container)

	log.Drop()

	args := docker.ContainerExecArgs{
		Name: name.Container,
		Cmd:  "dpkg-buildpackage" + " " + dpkgFlags,
	}
	err = dock.ContainerExec(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runTest() error {
	log.Info("Testing package")

	log.Drop()

	commands := []string{
		"debc",
		"sudo debi --with-depends",
		"lintian" + " " + lintianFlags,
	}

	for _, cmd := range commands {
		args := docker.ContainerExecArgs{
			Name: name.Container,
			Cmd:  cmd,
		}
		err := dock.ContainerExec(args)
		if err != nil {
			return log.FailE(err)
		}
	}

	return log.DoneE()
}

func runArchive() error {
	log.Info("Archiving build")

	info, err := os.Stat(name.ArchivePackageDir)
	if info != nil {
		err := os.RemoveAll(name.ArchivePackageDir)
		if err != nil {
			return log.FailE(err)
		}
	}

	err = os.Rename(name.BuildDir, name.ArchivePackageDir)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runScan() error {
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

func runStop() error {
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

func runRemove() error {
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
