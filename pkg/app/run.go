package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dawidd6/deber/pkg/stepping"

	"github.com/dawidd6/deber/pkg/logger"

	deb "github.com/dawidd6/deber/pkg/debian"
	doc "github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/spf13/cobra"
)

var (
	debian *deb.Debian
	docker *doc.Docker
	name   *naming.Naming
	log    *logger.Logger
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

	debian, err = initDebian(log)
	if err != nil {
		return err
	}

	docker, err = initDocker(log)
	if err != nil {
		return err
	}

	name = naming.New(
		cmd.Use,
		debian.TargetDist,
		debian.SourceName,
		debian.PackageVersion,
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

func initDocker(log *logger.Logger) (*doc.Docker, error) {
	log.Info("Connecting with Docker")

	docker, err := doc.New()
	if err != nil {
		return nil, log.FailE(err)
	}

	return docker, log.DoneE()
}

func initDebian(log *logger.Logger) (*deb.Debian, error) {
	log.Info("Parsing Debian changelog")

	debian, err := deb.ParseChangelog()
	if err != nil {
		return nil, log.FailE(err)
	}

	return debian, log.DoneE()
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

	err = docker.ExecShellContainer(name.Container)
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

	isImageBuilt, err := docker.IsImageBuilt(name.Image)
	if err != nil {
		return log.FailE(err)
	}
	if isImageBuilt {
		isImageOld, err := docker.IsImageOld(name.Image)
		if err != nil {
			return log.FailE(err)
		}
		if !isImageOld {
			return log.SkipE()
		}
	}

	for _, repo := range []string{"debian", "ubuntu"} {
		tags, err := doc.GetTags(repo)
		if err != nil {
			return log.FailE(err)
		}

		for _, tag := range tags {
			if tag.Name == debian.TargetDist {
				from := fmt.Sprintf("%s:%s", repo, debian.TargetDist)

				log.Drop()

				err := docker.BuildImage(name.Image, from)
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

	isContainerCreated, err := docker.IsContainerCreated(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	if isContainerCreated {
		return log.SkipE()
	}

	err = docker.CreateContainer(name)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runStart() error {
	log.Info("Starting container")

	isContainerStarted, err := docker.IsContainerStarted(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	if isContainerStarted {
		return log.SkipE()
	}

	err = docker.StartContainer(name.Container)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runTarball() error {
	log.Info("Moving tarball")

	tarball, err := debian.LocateTarball()
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

	err = docker.ExecContainer(name.Container, "sudo apt-get update")
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runDeps() error {
	log.Info("Installing dependencies")

	log.Drop()

	err := docker.ExecContainer(name.Container, "sudo mk-build-deps -ri")
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

	err = docker.DisableNetwork(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	defer docker.EnableNetwork(name.Container)

	log.Drop()

	err = docker.ExecContainer(name.Container, "dpkg-buildpackage"+" "+dpkgFlags)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runTest() error {
	log.Info("Testing package")

	log.Drop()

	err := docker.ExecContainer(name.Container, "debc")
	if err != nil {
		return log.FailE(err)
	}

	err = docker.ExecContainer(name.Container, "sudo debi --with-depends")
	if err != nil {
		return log.FailE(err)
	}

	err = docker.ExecContainer(name.Container, "lintian"+" "+lintianFlags)
	if err != nil {
		return log.FailE(err)
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

	err := docker.ExecContainer(name.Container, "cd "+naming.ContainerArchiveDir+" && dpkg-scanpackages -m . > Packages")
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runStop() error {
	log.Info("Stopping container")

	isContainerStopped, err := docker.IsContainerStopped(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	if isContainerStopped {
		return log.SkipE()
	}

	err = docker.StopContainer(name.Container)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func runRemove() error {
	log.Info("Removing container")

	isContainerCreated, err := docker.IsContainerCreated(name.Container)
	if err != nil {
		return log.FailE(err)
	}
	if !isContainerCreated {
		return log.SkipE()
	}

	err = docker.RemoveContainer(name.Container)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}
