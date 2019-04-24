package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	log = logger.New(cmd.Use)

	steps := map[string]func() error{
		"check":   runCheck,
		"build":   runBuild,
		"create":  runCreate,
		"start":   runStart,
		"tarball": runTarball,
		"update":  runUpdate,
		"deps":    runDeps,
		"package": runPackage,
		"test":    runTest,
		"archive": runArchive,
		"scan":    runScan,
		"stop":    runStop,
		"remove":  runRemove,
	}
	keys := []string{
		"check",
		"build",
		"create",
		"start",
		"tarball",
		"update",
		"deps",
		"package",
		"test",
		"archive",
		"scan",
		"stop",
		"remove",
	}

	log.Info("Parsing Debian changelog")
	debian, err = deb.ParseChangelog()
	if err != nil {
		return log.FailE(err)
	}
	log.Done()

	log.Info("Connecting with Docker")
	docker, err = doc.New()
	if err != nil {
		return log.FailE(err)
	}
	log.Done()

	name = naming.New(
		cmd.Use,
		debian.TargetDist,
		debian.SourceName,
		debian.PackageVersion,
		archiveDir,
	)

	if include != "" && exclude != "" {
		return errors.New("can't specify --include and --exclude together")
	}

	if include != "" {
		for key := range steps {
			if !strings.Contains(include, key) {
				delete(steps, key)
			}
		}
	}

	if exclude != "" {
		for key := range steps {
			if strings.Contains(exclude, key) {
				delete(steps, key)
			}
		}
	}

	for i := range keys {
		function, ok := steps[keys[i]]
		if ok {
			err := function()
			if err != nil {
				return err
			}
		}
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

	err := docker.ExecContainer(name.Container, "sudo mk-build-deps -ri -t apty")
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

	err = docker.ExecContainer(name.Container, "sudo debi --with-depends --tool apty")
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
