package cli

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
			Description: []string{
				"Checks if to-be-built package is already built and in archive.",
				"If package is in archive, then deber will simply exit.",
				"To build package anyway, simply exclude this step.",
			},
		}, {
			Name: "build",
			Run:  runBuild,
			Description: []string{
				"Builds image for deber's use.",
				"This step is skipped if an image is already built.",
				"Image's parent name is derived from Debian's changelog, for example",
				"if in `debian/changelog` target distribution is `bionic`, then",
				"deber will use `ubuntu:bionic` image as a parent from Docker Hub.",
				"Image's repository name is determined by querying Docker Hub API.",
				"So, if one wants to build for other distribution than specified in",
				"`debian/changelog`, just change target distribution to whatever",
				"one desires and deber will follow.",
				"Also if image is older than 14 days, deber will try to rebuild it.",
			},
		}, {
			Name: "create",
			Run:  runCreate,
			Description: []string{
				"Creates container and makes needed directories on host system.",
				"Will fail if image is nonexistent.",
			},
		}, {
			Name: "start",
			Run:  runStart,
			Description: []string{
				"Starts previously created container.",
				"The entry command is `sleep inf`, which means that container",
				"will just sit there, doing nothing and waiting for commands.",
			},
		}, {
			Name: "tarball",
			Run:  runTarball,
			Description: []string{
				"Moves orig upstream tarball from parent directory to build directory.",
				"Will fail if tarball is nonexistent and skip if package is native.",
			},
		}, {
			Name: "update",
			Run:  runUpdate,
			Description: []string{
				"Updates apt's cache.",
				"Also creates empty `Packages` file in archive if nonexistent",
			},
		}, {
			Name: "deps",
			Run:  runDeps,
			Description: []string{
				"Installs package's build dependencies in container.",
				"Runs `mk-build-deps` with appropriate options.",
			},
		}, {
			Name: "package",
			Run:  runPackage,
			Description: []string{
				"Runs `dpkg-buildpackage` in container.",
				"Options passed to `dpkg-buildpackage` are taken from environment variable",
				"Current `dpkg-buildpackage` options: " + dpkgFlags,
			},
		}, {
			Name: "test",
			Run:  runTest,
			Description: []string{
				"Runs series of commands in container:",
				"  - debc",
				"  - debi",
				"  - lintian",
				"Options passed to `lintian` are taken from environment variable",
				"Current `lintian` options: " + lintianFlags,
			},
		}, {
			Name: "archive",
			Run:  runArchive,
			Description: []string{
				"Moves built package artifacts (like .deb, .dsc and others) to archive.",
				"Package directory in archive is overwritten every time.",
			},
		}, {
			Name: "scan",
			Run:  runScan,
			Description: []string{
				"Scans available packages in archive and writes result to `Packages` file.",
				"This `Packages` file is then used by apt in container.",
			},
		}, {
			Name: "stop",
			Run:  runStop,
			Description: []string{
				"Stops container.",
				"With " + docker.ContainerStopTimeout.String() + " timeout.",
			},
		}, {
			Name: "remove",
			Run:  runRemove,
			Description: []string{
				"Removes container.",
				"Nothing more.",
			},
		},
	}

	// SECTION: handle options
	switch {
	case shell:
		steps.Reset()
		steps.ExtraFunctionAfterRun(runShellOptional)

		err := steps.Include("build", "create", "start")
		if err != nil {
			return err
		}
	case remove:
		steps.Reset()

		err := steps.Include("remove", "stop")
		if err != nil {
			return err
		}
	case list:
		for i, step := range steps {
			fmt.Printf("%d. %s\n\n", i+1, step.Name)
			for _, desc := range step.Description {
				fmt.Printf("\t%s\n", desc)
			}

			if i < len(steps)-1 {
				fmt.Println()
			}
		}
		return nil
	default:
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

	// SECTION: run steps
	err = steps.Run()
	if err != nil {
		return err
	}

	return nil
}

func runShellOptional() error {
	args := docker.ContainerExecArgs{
		Interactive: true,
		Name:        name.Container,
	}
	err := dock.ContainerExec(args)
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
	info, _ := os.Stat(file)
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
	err := dock.ContainerExec(args)
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
	info, _ := os.Stat(file)
	if info == nil {
		_, err := os.Create(file)
		if err != nil {
			return log.FailE(err)
		}
	}

	err := dock.ContainerDisableNetwork(name.Container)
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
