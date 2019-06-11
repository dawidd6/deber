package steps

import (
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/docker/docker/api/types/mount"
	"os"
	"path/filepath"
	"strings"
)

var (
	DpkgFlags     = "-tc"
	LintianFlags  = "-i -I"
	NoRebuild     bool
	ExtraPackages []string
)

func Steps() []func(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	return []func(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error{
		RunBuild,
		RunCreate,
		RunStart,
		RunTarball,
		RunDepends,
		RunPackage,
		RunTest,
		RunArchive,
		RunStop,
		RunRemove,
	}
}

// RunBuild function determines parent image name by querying DockerHub API
// for available "debian" and "ubuntu" tags and confronting them with
// debian/changelog's target distribution.
//
// At last it commands Docker Engine to build image.
func RunBuild(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Building image")

	isImageBuilt, err := dock.IsImageBuilt(n.ImageName())
	if err != nil {
		return log.Result(err)
	}
	if isImageBuilt {
		isImageOld, err := dock.IsImageOld(n.ImageName())
		if err != nil {
			return log.Result(err)
		}
		if !isImageOld {
			return log.Result(log.Skip)
		} else if NoRebuild {
			return log.Result(log.Skip)
		}
	}

	repos := []string{"debian", "ubuntu"}
	repo, err := docker.MatchRepo(repos, n.ImageTag())
	if err != nil {
		return log.Result(err)
	}

	log.Drop()

	args := docker.ImageBuildArgs{
		From: fmt.Sprintf("%s:%s", repo, n.ImageTag()),
		Name: n.ImageName(),
	}
	err = dock.ImageBuild(args)
	if err != nil {
		return log.Result(err)
	}

	return log.Result(nil)
}

// RunCreate function commands Docker Engine to create container.
func RunCreate(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Creating container")

	isContainerCreated, err := dock.IsContainerCreated(n.ContainerName())
	if err != nil {
		return log.Result(err)
	}
	if isContainerCreated {
		return log.Result(log.Skip)
	}
	//TODO check if mounts are equal

	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: n.SourceDir(),
			Target: docker.ContainerSourceDir,
		}, {
			Type:   mount.TypeBind,
			Source: n.BuildDir(),
			Target: docker.ContainerBuildDir,
		}, {
			Type:   mount.TypeBind,
			Source: n.CacheDir(),
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
			return log.Result(err)
		}
	}

	for _, pkg := range ExtraPackages {
		source, err := filepath.Abs(pkg)
		if err != nil {
			return log.Result(err)
		}

		info, err := os.Stat(source)
		if info == nil {
			return log.Result(err)
		}
		if !info.IsDir() && !strings.HasSuffix(source, ".deb") {
			return log.Result(errors.New("please specify a directory or .deb file"))
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
		Image:  n.ImageName(),
		Name:   n.ContainerName(),
		User:   fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
	}
	err = dock.ContainerCreate(args)
	if err != nil {
		return log.Result(err)
	}

	return log.Result(nil)
}

// RunStart function commands Docker Engine to start container.
func RunStart(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Starting container")

	isContainerStarted, err := dock.IsContainerStarted(n.ContainerName())
	if err != nil {
		return log.Result(err)
	}
	if isContainerStarted {
		return log.Result(log.Skip)
	}

	err = dock.ContainerStart(n.ContainerName())
	if err != nil {
		return log.Result(err)
	}

	return log.Result(nil)
}

// RunTarball function moves orig upstream tarball from parent directory
// to build directory if package is not native.
func RunTarball(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Moving tarball")

	if deb.Version.Native {
		return log.Result(log.Skip)
	}

	// Skip if tarball is already in build directory.
	tarball, found := deb.FindTarball(n.BuildDir())
	if found {
		return log.Result(log.Skip)
	}

	tarball, found = deb.FindTarball(n.SourceParentDir())
	if !found {
		return log.Result(errors.New("tarball not found"))
	}

	source := filepath.Join(n.SourceParentDir(), tarball)
	dest := filepath.Join(n.BuildDir(), tarball)

	source, err := filepath.EvalSymlinks(source)
	if err != nil {
		return log.Result(err)
	}

	err = os.Rename(source, dest)
	if err != nil {
		return log.Result(err)
	}

	return log.Result(nil)
}

func RunDepends(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Installing dependencies")
	log.Drop()

	args := []docker.ContainerExecArgs{
		{
			Name:    n.ContainerName(),
			Cmd:     "rm -f a.list",
			AsRoot:  true,
			WorkDir: "/etc/apt/sources.list.d",
		}, {
			Name:    n.ContainerName(),
			Cmd:     "echo deb [trusted=yes] file://" + docker.ContainerArchiveDir + " ./ > a.list",
			AsRoot:  true,
			WorkDir: "/etc/apt/sources.list.d",
		}, {
			Name:    n.ContainerName(),
			Cmd:     "dpkg-scanpackages -m . > Packages",
			AsRoot:  true,
			WorkDir: docker.ContainerArchiveDir,
		}, {
			Name:    n.ContainerName(),
			Cmd:     "apt-get update",
			AsRoot:  true,
			Network: true,
		}, {
			Name:    n.ContainerName(),
			Cmd:     "apt-get build-dep ./",
			Network: true,
			AsRoot:  true,
		},
	}

	if ExtraPackages == nil {
		args[1].Skip = true
		args[2].Skip = true
	}

	for _, arg := range args {
		err := dock.ContainerExec(arg)
		if err != nil {
			return log.Result(err)
		}
	}

	return log.Result(nil)
}

// RunPackage function first disables network in container,
// then executes "dpkg-buildpackage" and at the end,
// enables network back.
func RunPackage(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Packaging software")
	log.Drop()

	args := docker.ContainerExecArgs{
		Name: n.ContainerName(),
		Cmd:  "dpkg-buildpackage" + " " + DpkgFlags,
	}
	err := dock.ContainerExec(args)
	if err != nil {
		return log.Result(err)
	}

	return log.Result(nil)
}

// RunTest function executes "debc", "debi" and "lintian" in container.
func RunTest(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Testing package")
	log.Drop()

	args := []docker.ContainerExecArgs{
		{
			Name: n.ContainerName(),
			Cmd:  "debc",
		}, {
			Name:    n.ContainerName(),
			Cmd:     "debi --with-depends",
			Network: true,
			AsRoot:  true,
		}, {
			Name: n.ContainerName(),
			Cmd:  "lintian" + " " + LintianFlags,
		},
	}

	for _, arg := range args {
		err := dock.ContainerExec(arg)
		if err != nil {
			return log.Result(err)
		}
	}

	return log.Result(nil)
}

// RunArchive function moves successful build to archive by overwriting.
func RunArchive(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Archiving build")

	err := os.MkdirAll(n.ArchiveSourceDir(), os.ModePerm)
	if err != nil {
		return log.Result(err)
	}

	info, _ := os.Stat(n.ArchiveVersionDir())
	if info != nil {
		err := os.RemoveAll(n.ArchiveVersionDir())
		if err != nil {
			return log.Result(err)
		}
	}

	err = os.Rename(n.BuildDir(), n.ArchiveVersionDir())
	if err != nil {
		return log.Result(err)
	}

	return log.Result(nil)
}

// RunStop function commands Docker Engine to stop container.
func RunStop(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Stopping container")

	isContainerStopped, err := dock.IsContainerStopped(n.ContainerName())
	if err != nil {
		return log.Result(err)
	}
	if isContainerStopped {
		return log.Result(log.Skip)
	}

	err = dock.ContainerStop(n.ContainerName())
	if err != nil {
		return log.Result(err)
	}

	return log.Result(nil)
}

// RunRemove function commands Docker Engine to remove container.
func RunRemove(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Removing container")

	isContainerCreated, err := dock.IsContainerCreated(n.ContainerName())
	if err != nil {
		return log.Result(err)
	}
	if !isContainerCreated {
		return log.Result(log.Skip)
	}

	err = dock.ContainerRemove(n.ContainerName())
	if err != nil {
		return log.Result(err)
	}

	return log.Result(nil)
}

// RunShellOptional function interactively executes bash shell in container.
func RunShellOptional(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Launching shell")

	args := docker.ContainerExecArgs{
		Interactive: true,
		AsRoot:      true,
		Name:        n.ContainerName(),
	}
	err := dock.ContainerExec(args)
	if err != nil {
		return log.Result(err)
	}

	return log.Result(nil)
}

// RunCheck function evaluates if package has been already built and
// is in archive, if it is, then it exits with 0 code.
func RunCheckOptional(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Checking archive")

	info, _ := os.Stat(n.ArchiveVersionDir())
	if info != nil {
		os.Exit(0)
	}

	return log.Result(nil)
}

func RunInfoOptional(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	fmt.Println("Debian:")
	fmt.Printf("  DpkgFlags = %s\n", DpkgFlags)
	fmt.Printf("  LintianFlags = %s\n", LintianFlags)
	fmt.Printf("  DebianSource = %s\n", deb.Source)
	fmt.Printf("  DebianPackageVersion = %s\n", deb.Version.Package)
	fmt.Printf("  DebianPackageUpstream = %s\n", deb.Version.Upstream)
	fmt.Printf("  DebianTarget = %s\n", deb.Target)

	fmt.Println()

	fmt.Println("Docker:")
	fmt.Printf("  Image = %s\n", n.ImageName())
	fmt.Printf("  Container = %s\n", n.ContainerName())
	fmt.Printf("  ArchiveSourceDir = %s\n", n.ArchiveSourceDir())
	fmt.Printf("  BuildDir = %s\n", n.BuildDir())
	fmt.Printf("  CacheDir = %s\n", n.CacheDir())
	fmt.Printf("  SourceDir = %s\n", n.SourceDir())

	return nil
}
