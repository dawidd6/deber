package steps

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/dockerfile"
	"github.com/dawidd6/deber/pkg/dockerhub"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/docker/docker/api/types/mount"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	DpkgFlags     = "-tc"
	LintianFlags  = "-i -I"
	NoRebuild     bool
	NoUpdate      bool
	WithNetwork   bool
	ExtraPackages []string
)

// Build function determines parent image name by querying DockerHub API
// for available "debian" and "ubuntu" tags and confronting them with
// debian/changelog's target distribution.
//
// At last it commands Docker Engine to build image.
func Build(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Building image")

	isImageBuilt, err := dock.IsImageBuilt(n.ImageName())
	if err != nil {
		return log.Failed(err)
	}
	if isImageBuilt {
		isImageOld, err := dock.IsImageOld(n.ImageName())
		if err != nil {
			return log.Failed(err)
		}
		if !isImageOld {
			return log.Skipped()
		} else if NoRebuild {
			return log.Skipped()
		}
	}

	tag := strings.Split(n.ImageName(), ":")[1]
	repos := []string{"debian", "ubuntu"}
	repo, err := dockerhub.MatchRepo(repos, tag)
	if err != nil {
		return log.Failed(err)
	}

	file, err := dockerfile.Parse(repo, tag)
	if err != nil {
		return log.Failed(err)
	}

	log.Drop()

	args := docker.ImageBuildArgs{
		Name:       n.ImageName(),
		Dockerfile: file,
	}
	err = dock.ImageBuild(args)
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

// Create function commands Docker Engine to create container.
func Create(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Creating container")

	isContainerCreated, err := dock.IsContainerCreated(n.ContainerName())
	if err != nil {
		return log.Failed(err)
	}
	if isContainerCreated {
		return log.Skipped()
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
			return log.Failed(err)
		}
	}

	for _, pkg := range ExtraPackages {
		source, err := filepath.Abs(pkg)
		if err != nil {
			return log.Failed(err)
		}

		info, err := os.Stat(source)
		if info == nil {
			return log.Failed(err)
		}
		if !info.IsDir() && !strings.HasSuffix(source, ".deb") {
			return log.Failed(errors.New("please specify a directory or .deb file"))
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
		return log.Failed(err)
	}

	return log.Done()
}

// Start function commands Docker Engine to start container.
func Start(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Starting container")

	isContainerStarted, err := dock.IsContainerStarted(n.ContainerName())
	if err != nil {
		return log.Failed(err)
	}
	if isContainerStarted {
		return log.Skipped()
	}

	err = dock.ContainerStart(n.ContainerName())
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

// Tarball function moves orig upstream tarball from parent directory
// to build directory if package is not native.
func Tarball(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Moving tarball")

	if deb.Version.Native {
		return log.Skipped()
	}

	// Skip if tarball is already in build directory.
	tarball, found := deb.FindTarball(n.BuildDir())
	if found {
		return log.Skipped()
	}

	tarball, found = deb.FindTarball(n.SourceParentDir())
	if !found {
		return log.Failed(errors.New("tarball not found"))
	}

	source := filepath.Join(n.SourceParentDir(), tarball)
	dest := filepath.Join(n.BuildDir(), tarball)

	source, err := filepath.EvalSymlinks(source)
	if err != nil {
		return log.Failed(err)
	}

	err = os.Rename(source, dest)
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

func Depends(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
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
			Skip:    NoUpdate,
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
			return log.Failed(err)
		}
	}

	return log.Done()
}

// Package function first disables network in container,
// then executes "dpkg-buildpackage" and at the end,
// enables network back.
func Package(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Packaging software")
	log.Drop()

	args := docker.ContainerExecArgs{
		Name:    n.ContainerName(),
		Cmd:     "dpkg-buildpackage" + " " + DpkgFlags,
		Network: WithNetwork,
	}
	err := dock.ContainerExec(args)
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

// Test function executes "debi", "debc" and "lintian" in container.
func Test(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Testing package")
	log.Drop()

	args := []docker.ContainerExecArgs{
		{
			Name:    n.ContainerName(),
			Cmd:     "debi --with-depends",
			Network: true,
			AsRoot:  true,
		}, {
			Name: n.ContainerName(),
			Cmd:  "debc",
		}, {
			Name: n.ContainerName(),
			Cmd:  "lintian" + " " + LintianFlags,
		},
	}

	for _, arg := range args {
		err := dock.ContainerExec(arg)
		if err != nil {
			return log.Failed(err)
		}
	}

	return log.Done()
}

// Archive function moves successful build to archive by overwriting.
func Archive(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Archiving build")
	log.Drop()

	err := os.MkdirAll(n.ArchiveVersionDir(), os.ModePerm)
	if err != nil {
		return log.Failed(err)
	}

	files, err := ioutil.ReadDir(n.BuildDir())
	if err != nil {
		return log.Failed(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		log.ExtraInfo(file.Name())

		sourcePath := filepath.Join(n.BuildDir(), file.Name())
		targetPath := filepath.Join(n.ArchiveVersionDir(), file.Name())

		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			return log.Failed(err)
		}

		sourceBytes, err := ioutil.ReadAll(sourceFile)
		if err != nil {
			return log.Failed(err)
		}

		targetStat, _ := os.Stat(targetPath)
		if targetStat != nil {
			targetFile, err := os.Open(targetPath)
			if err != nil {
				return log.Failed(err)
			}

			targetBytes, err := ioutil.ReadAll(targetFile)
			if err != nil {
				return log.Failed(err)
			}

			sourceChecksum := md5.Sum(sourceBytes)
			targetChecksum := md5.Sum(targetBytes)

			if targetChecksum == sourceChecksum {
				_ = log.Skipped()
				continue
			}
		}

		targetFile, err := os.Create(targetPath)
		if err != nil {
			return log.Failed(err)
		}

		_, err = io.Copy(targetFile, sourceFile)
		if err != nil {
			return log.Failed(err)
		}

		_ = log.Done()
	}

	return log.None()
}

// Stop function commands Docker Engine to stop container.
func Stop(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Stopping container")

	isContainerStopped, err := dock.IsContainerStopped(n.ContainerName())
	if err != nil {
		return log.Failed(err)
	}
	if isContainerStopped {
		return log.Skipped()
	}

	err = dock.ContainerStop(n.ContainerName())
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

// Remove function commands Docker Engine to remove container.
func Remove(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Removing container")

	isContainerCreated, err := dock.IsContainerCreated(n.ContainerName())
	if err != nil {
		return log.Failed(err)
	}
	if !isContainerCreated {
		return log.Skipped()
	}

	err = dock.ContainerRemove(n.ContainerName())
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

// ShellOptional function interactively executes bash shell in container.
func ShellOptional(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Launching shell")

	args := docker.ContainerExecArgs{
		Interactive: true,
		AsRoot:      true,
		Name:        n.ContainerName(),
	}
	err := dock.ContainerExec(args)
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

// Check function evaluates if package has been already built and
// is in archive, if it is, then it exits with 0 code.
func CheckOptional(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
	log.Info("Checking archive")

	info, _ := os.Stat(n.ArchiveVersionDir())
	if info != nil {
		_ = log.Skipped()
		os.Exit(0)
	}

	return log.Done()
}

func InfoOptional(dock *docker.Docker, deb *debian.Debian, n *naming.Naming) error {
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
