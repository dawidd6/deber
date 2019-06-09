package app

import (
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/util"
	"github.com/docker/docker/api/types/mount"
	"os"
	"path/filepath"
	"strings"
)

func (a *App) Steps() []func() error {
	return []func() error{
		a.RunBuild,
		a.RunCreate,
		a.RunStart,
		a.RunTarball,
		a.RunDepends,
		a.RunPackage,
		a.RunTest,
		a.RunArchive,
		a.RunStop,
		a.RunRemove,
	}
}

// RunBuild function determines parent image name by querying DockerHub API
// for available "debian" and "ubuntu" tags and confronting them with
// debian/changelog's target distribution.
//
// At last it commands Docker Engine to build image.
func (a *App) RunBuild() error {
	a.LogInfo("Building image")

	isImageBuilt, err := a.Docker.IsImageBuilt(a.ImageName())
	if err != nil {
		return a.LogResult(err)
	}
	if isImageBuilt {
		isImageOld, err := a.Docker.IsImageOld(a.ImageName())
		if err != nil {
			return a.LogResult(err)
		}
		if !isImageOld {
			return a.LogResult(logSkip)
		} else if a.Config.NoRebuild {
			return a.LogResult(logSkip)
		}
	}

	repos := []string{"debian", "ubuntu"}
	repo, err := docker.MatchRepo(repos, a.ImageTag())
	if err != nil {
		return a.LogResult(err)
	}

	a.LogDrop()

	args := docker.ImageBuildArgs{
		From: fmt.Sprintf("%s:%s", repo, a.ImageTag()),
		Name: a.ImageName(),
	}
	err = a.Docker.ImageBuild(args)
	if err != nil {
		return a.LogResult(err)
	}

	return a.LogResult(nil)
}

// RunCreate function commands Docker Engine to create container.
func (a *App) RunCreate() error {
	a.LogInfo("Creating container")

	isContainerCreated, err := a.Docker.IsContainerCreated(a.ContainerName())
	if err != nil {
		return a.LogResult(err)
	}
	if isContainerCreated {
		return a.LogResult(logSkip)
	}
	//TODO check if mounts are equal

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
			return a.LogResult(err)
		}
	}

	for _, pkg := range a.Config.ExtraPackages {
		source, err := filepath.Abs(pkg)
		if err != nil {
			return a.LogResult(err)
		}

		info, err := os.Stat(source)
		if info == nil {
			return a.LogResult(err)
		}
		if !info.IsDir() && !strings.HasSuffix(source, ".deb") {
			return a.LogResult(errors.New("please specify a directory or .deb file"))
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
	err = a.Docker.ContainerCreate(args)
	if err != nil {
		return a.LogResult(err)
	}

	return a.LogResult(nil)
}

// RunStart function commands Docker Engine to start container.
func (a *App) RunStart() error {
	a.LogInfo("Starting container")

	isContainerStarted, err := a.Docker.IsContainerStarted(a.ContainerName())
	if err != nil {
		return a.LogResult(err)
	}
	if isContainerStarted {
		return a.LogResult(logSkip)
	}

	err = a.Docker.ContainerStart(a.ContainerName())
	if err != nil {
		return a.LogResult(err)
	}

	return a.LogResult(nil)
}

// RunTarball function moves orig upstream tarball from parent directory
// to build directory if package is not native.
func (a *App) RunTarball() error {
	a.LogInfo("Moving tarball")

	if a.Debian.Version.IsNative() {
		return a.LogResult(logSkip)
	}

	// Skip if tarball is already in build directory.
	tarball, found := util.FindTarball(a.Debian, a.BuildDir())
	if found {
		return a.LogResult(logSkip)
	}

	tarball, found = util.FindTarball(a.Debian, a.SourceParentDir())
	if !found {
		return a.LogResult(errors.New("tarball not found"))
	}

	source := filepath.Join(a.SourceParentDir(), tarball)
	dest := filepath.Join(a.BuildDir(), tarball)

	source, err := filepath.EvalSymlinks(source)
	if err != nil {
		return a.LogResult(err)
	}

	return os.Rename(source, dest)
}

func (a *App) RunDepends() error {
	a.LogInfo("Installing dependencies")
	a.LogDrop()

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

	if a.Config.ExtraPackages != nil {
		args = append(args, extraPreArgs...)
	}
	args = append(args, standardArgs...)

	for _, arg := range args {
		err := a.Docker.ContainerExec(arg)
		if err != nil {
			return a.LogResult(err)
		}
	}

	return a.LogResult(nil)
}

// RunPackage function first disables network in container,
// then executes "dpkg-buildpackage" and at the end,
// enables network back.
func (a *App) RunPackage() error {
	a.LogInfo("Packaging software")
	a.LogDrop()

	args := docker.ContainerExecArgs{
		Name: a.ContainerName(),
		Cmd:  "dpkg-buildpackage" + " " + a.Config.DpkgFlags,
	}
	err := a.Docker.ContainerExec(args)
	if err != nil {
		return a.LogResult(err)
	}

	return a.LogResult(nil)
}

// RunTest function executes "debc", "debi" and "lintian" in container.
func (a *App) RunTest() error {
	a.LogInfo("Testing package")
	a.LogDrop()

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
			Cmd:  "lintian" + " " + a.Config.LintianFlags,
		},
	}

	for _, arg := range args {
		err := a.Docker.ContainerExec(arg)
		if err != nil {
			return a.LogResult(err)
		}
	}

	return a.LogResult(nil)
}

// RunArchive function moves successful build to archive by overwriting.
func (a *App) RunArchive() error {
	a.LogInfo("Archiving build")

	err := os.MkdirAll(a.ArchiveSourceDir(), os.ModePerm)
	if err != nil {
		return a.LogResult(err)
	}

	info, _ := os.Stat(a.ArchiveVersionDir())
	if info != nil {
		err := os.RemoveAll(a.ArchiveVersionDir())
		if err != nil {
			return a.LogResult(err)
		}
	}

	err = os.Rename(a.BuildDir(), a.ArchiveVersionDir())
	if err != nil {
		return a.LogResult(err)
	}

	return a.LogResult(nil)
}

// RunStop function commands Docker Engine to stop container.
func (a *App) RunStop() error {
	a.LogInfo("Stopping container")

	isContainerStopped, err := a.Docker.IsContainerStopped(a.ContainerName())
	if err != nil {
		return a.LogResult(err)
	}
	if isContainerStopped {
		return a.LogResult(logSkip)
	}

	err = a.Docker.ContainerStop(a.ContainerName())
	if err != nil {
		return a.LogResult(err)
	}

	return a.LogResult(nil)
}

// RunRemove function commands Docker Engine to remove container.
func (a *App) RunRemove() error {
	a.LogInfo("Removing container")

	isContainerCreated, err := a.Docker.IsContainerCreated(a.ContainerName())
	if err != nil {
		return a.LogResult(err)
	}
	if !isContainerCreated {
		return a.LogResult(logSkip)
	}

	err = a.Docker.ContainerRemove(a.ContainerName())
	if err != nil {
		return a.LogResult(err)
	}

	return a.LogResult(nil)
}

// RunShellOptional function interactively executes bash shell in container.
func (a *App) RunShellOptional() error {
	a.LogInfo("Launching shell")

	args := docker.ContainerExecArgs{
		Interactive: true,
		AsRoot:      true,
		Name:        a.ContainerName(),
	}
	err := a.Docker.ContainerExec(args)
	if err != nil {
		return a.LogResult(err)
	}

	return a.LogResult(nil)
}

// RunCheck function evaluates if package has been already built and
// is in archive, if it is, then it exits with 0 code.
func (a *App) RunCheckOptional() error {
	a.LogInfo("Checking archive")

	info, _ := os.Stat(a.ArchiveVersionDir())
	if info != nil {
		os.Exit(0)
	}

	return a.LogResult(nil)
}
