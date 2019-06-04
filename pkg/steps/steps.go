package steps

import (
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/util"
	"github.com/docker/docker/api/types/mount"
	"os"
	"path/filepath"
	"strings"
)

func Run(a *app.App) error {
	steps := []func(*app.App) error{
		runBuild,
		runCreate,
		runStart,
		runTarball,
		runDepends,
		runPackage,
		runTest,
		runArchive,
		runStop,
		runRemove,
	}

	for _, step := range steps {
		err := step(a)
		if err != nil {
			return err
		}
	}

	return nil
}

// runBuild function determines parent image name by querying DockerHub API
// for available "debian" and "ubuntu" tags and confronting them with
// debian/changelog's target distribution.
//
// At last it commands Docker Engine to build image.
func runBuild(a *app.App) error {
	var err error
	a.Info("Building image")

	isImageBuilt, err := a.IsImageBuilt(a.ImageName())
	if err != nil {
		return err
	}
	if isImageBuilt {
		isImageOld, err := a.IsImageOld(a.ImageName())
		if err != nil {
			return err
		}
		if !isImageOld {
			return nil
		}
	}

	repos := []string{"debian", "ubuntu"}
	repo, err := docker.MatchRepo(repos, a.ImageTag())
	if err != nil {
		return err
	}

	args := docker.ImageBuildArgs{
		From: fmt.Sprintf("%s:%s", repo, a.ImageTag()),
		Name: a.ImageName(),
	}
	err = a.ImageBuild(args)
	if err != nil {
		return err
	}

	return nil
}

// runCreate function commands Docker Engine to create container.
func runCreate(a *app.App) error {
	a.Info("Creating container")

	isContainerCreated, err := a.IsContainerCreated(a.ContainerName())
	if err != nil {
		return err
	}
	if isContainerCreated {
		return nil
	}

	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: a.SourceDir(),
			Target: docker.ContainerSourceDir,
		}, {
			Type:   mount.TypeBind,
			Source: a.BuildDir(),
			Target: docker.ContainerBuildDir,
		}, {
			Type:   mount.TypeBind,
			Source: a.CacheDir(),
			Target: docker.ContainerCacheDir,
		},
	}

	for _, mnt := range mounts {
		info, _ := os.Stat(mnt.Source)
		if info != nil {
			continue
		}

		err := os.MkdirAll(mnt.Source, os.ModePerm)
		if err != nil {
			return err
		}
	}

	for _, pkg := range a.ExtraPackages {
		source, err := filepath.Abs(pkg)
		if err != nil {
			return err
		}

		info, err := os.Stat(source)
		if info == nil {
			return err
		}
		if !info.IsDir() && !strings.HasSuffix(source, ".deb") {
			return errors.New("please specify a directory or .deb file")
		}

		target := filepath.Join(docker.ContainerArchiveDir, filepath.Base(source))

		mnt := mount.Mount{
			Type:     mount.TypeBind,
			Source:   source,
			Target:   target,
			ReadOnly: true,
		}

		mounts = append(mounts, mnt)
	}

	args := docker.ContainerCreateArgs{
		Mounts: mounts,
		Image:  a.ImageName(),
		Name:   a.ContainerName(),
		User:   fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
	}
	err = a.ContainerCreate(args)
	if err != nil {
		return err
	}

	return nil
}

// runStart function commands Docker Engine to start container.
func runStart(a *app.App) error {
	a.Info("Starting container")

	isContainerStarted, err := a.IsContainerStarted(a.ContainerName())
	if err != nil {
		return err
	}
	if isContainerStarted {
		return nil
	}

	err = a.ContainerStart(a.ContainerName())
	if err != nil {
		return err
	}

	return nil
}

// runTarball function moves orig upstream tarball from parent directory
// to build directory if package is not native.
func runTarball(a *app.App) error {
	a.Info("Moving tarball")

	file, dir, err := util.FindTarball(a)
	if err != nil {
		return err
	}

	source := filepath.Join(dir, file)
	source, err = filepath.EvalSymlinks(source)
	if err != nil {
		return err
	}

	dest := filepath.Join(a.BuildDir(), file)
	err = os.Rename(source, dest)
	if err != nil {
		return err
	}

	return nil
}

func runDepends(a *app.App) error {
	a.Info("Installing dependencies")

	args := make([]docker.ContainerExecArgs, 0)

	extraPreArgs := []docker.ContainerExecArgs{
		{
			Name:    a.ContainerName(),
			Cmd:     "echo deb [trusted=yes] file://" + docker.ContainerArchiveDir + " ./ > a.list",
			AsRoot:  true,
			WorkDir: "/etc/apt/sources.list.d",
		}, {
			Name:    a.ContainerName(),
			Cmd:     "dpkg-scanpackages -m . > Packages",
			AsRoot:  true,
			WorkDir: docker.ContainerArchiveDir,
		},
	}

	standardArgs := []docker.ContainerExecArgs{
		{
			Name:    a.ContainerName(),
			Cmd:     "apt-get update",
			AsRoot:  true,
			Network: true,
		}, {
			Name:    a.ContainerName(),
			Cmd:     "apt-get build-dep ./",
			Network: true,
			AsRoot:  true,
		},
	}

	if a.ExtraPackages != nil {
		args = append(args, extraPreArgs...)
	}
	args = append(args, standardArgs...)

	for _, arg := range args {
		err := a.ContainerExec(arg)
		if err != nil {
			return err
		}
	}

	return nil
}

// runPackage function first disables network in container,
// then executes "dpkg-buildpackage" and at the end,
// enables network back.
func runPackage(a *app.App) error {
	a.Info("Packaging software")

	args := docker.ContainerExecArgs{
		Name: a.ContainerName(),
		Cmd:  "dpkg-buildpackage" + " " + a.DpkgFlags,
	}
	err := a.ContainerExec(args)
	if err != nil {
		return err
	}

	return nil
}

// runTest function executes "debc", "debi" and "lintian" in container.
func runTest(a *app.App) error {
	a.Info("Testing package")

	args := []docker.ContainerExecArgs{
		{
			Name: a.ContainerName(),
			Cmd:  "debc",
		}, {
			Name:    a.ContainerName(),
			Cmd:     "debi --with-depends",
			Network: true,
			AsRoot:  true,
		}, {
			Name: a.ContainerName(),
			Cmd:  "lintian" + " " + a.LintianFlags,
		},
	}

	for _, arg := range args {
		err := a.ContainerExec(arg)
		if err != nil {
			return err
		}
	}

	return nil
}

// runArchive function moves successful build to archive by overwriting.
func runArchive(a *app.App) error {
	a.Info("Archiving build")

	err := os.MkdirAll(a.ArchiveSourceDir(), os.ModePerm)
	if err != nil {
		return err
	}

	info, _ := os.Stat(a.ArchiveVersionDir())
	if info != nil {
		err := os.RemoveAll(a.ArchiveVersionDir())
		if err != nil {
			return err
		}
	}

	err = os.Rename(a.BuildDir(), a.ArchiveVersionDir())
	if err != nil {
		return err
	}

	return nil
}

// runStop function commands Docker Engine to stop container.
func runStop(a *app.App) error {
	a.Info("Stopping container")

	isContainerStopped, err := a.IsContainerStopped(a.ContainerName())
	if err != nil {
		return err
	}
	if isContainerStopped {
		return nil
	}

	err = a.ContainerStop(a.ContainerName())
	if err != nil {
		return err
	}

	return nil
}

// runRemove function commands Docker Engine to remove container.
func runRemove(a *app.App) error {
	a.Info("Removing container")

	isContainerCreated, err := a.IsContainerCreated(a.ContainerName())
	if err != nil {
		return err
	}
	if !isContainerCreated {
		return nil
	}

	err = a.ContainerRemove(a.ContainerName())
	if err != nil {
		return err
	}

	return nil
}

// runShellOptional function interactively executes bash shell in container.
func runShellOptional(a *app.App) error {
	a.Info("Launching shell")

	args := docker.ContainerExecArgs{
		Interactive: true,
		AsRoot:      true,
		Name:        a.ContainerName(),
	}
	err := a.ContainerExec(args)
	if err != nil {
		return err
	}

	return nil
}

// runCheck function evaluates if package has been already built and
// is in archive, if it is, then it exits with 0 code.
func runCheckOptional(a *app.App) error {
	a.Info("Checking archive")

	info, _ := os.Stat(a.ArchiveVersionDir())
	if info != nil {
		os.Exit(0)
	}

	return nil
}
