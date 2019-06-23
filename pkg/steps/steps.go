package steps

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/dockerfile"
	"github.com/dawidd6/deber/pkg/dockerhub"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/util"
	"github.com/dawidd6/deber/pkg/walk"
	"github.com/docker/docker/api/types/mount"
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
func Build() error {
	log.Info("Building image")

	isImageBuilt, err := docker.IsImageBuilt(naming.Image())
	if err != nil {
		return log.Failed(err)
	}
	if isImageBuilt {
		isImageOld, err := docker.IsImageOld(naming.Image())
		if err != nil {
			return log.Failed(err)
		}
		if !isImageOld {
			return log.Skipped()
		} else if NoRebuild {
			return log.Skipped()
		}
	}

	tag := strings.Split(naming.Image(), ":")[1]
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
		Name:       naming.Image(),
		Dockerfile: file,
	}
	err = docker.ImageBuild(args)
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

// Create function commands Docker Engine to create container.
//
// Also makes directories on host and moves tarball if needed.
func Create() error {
	log.Info("Creating container")

	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: naming.SourceDir(),
			Target: docker.ContainerSourceDir,
		}, {
			Type:   mount.TypeBind,
			Source: naming.BuildDir(),
			Target: docker.ContainerBuildDir,
		}, {
			Type:   mount.TypeBind,
			Source: naming.CacheDir(),
			Target: docker.ContainerCacheDir,
		},
	}

	// Handle extra packages mounting
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

	isContainerCreated, err := docker.IsContainerCreated(naming.Container())
	if err != nil {
		return log.Failed(err)
	}
	if isContainerCreated {
		oldMounts, err := docker.ContainerMounts(naming.Container())
		if err != nil {
			return log.Failed(err)
		}

		// Compare old mounts with new ones,
		// if not equal, then recreate container
		if util.CompareMounts(oldMounts, mounts) {
			return log.Skipped()
		}

		err = docker.ContainerStop(naming.Container())
		if err != nil {
			return log.Failed(err)
		}

		err = docker.ContainerRemove(naming.Container())
		if err != nil {
			return log.Failed(err)
		}
	}

	// Make directories if non existent
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

	// Find tarball if package is not native
	if strings.Contains(naming.PackageVersion, "-") {
		tarball := fmt.Sprintf("%s_%s.orig.tar", naming.PackageName, naming.PackageUpstream)
		found := false

		// Look for tarball in build directory,
		// if it's there, then do nothing
		files, err := ioutil.ReadDir(naming.BuildDir())
		if err != nil {
			return log.Failed(err)
		}

		for _, file := range files {
			if strings.HasPrefix(file.Name(), tarball) {
				tarball = file.Name()
				found = true
				break
			}
		}

		// If tarball is not present in build directory,
		// then look in parent source directory
		if !found {
			files, err := ioutil.ReadDir(naming.SourceParentDir())
			if err != nil {
				return log.Failed(err)
			}

			for _, file := range files {
				if strings.HasPrefix(file.Name(), tarball) {
					tarball = file.Name()
					found = true
					break
				}
			}

			if !found {
				return log.Failed(errors.New("upstream tarball not found"))
			}

			source := filepath.Join(naming.SourceParentDir(), tarball)
			dst := filepath.Join(naming.BuildDir(), tarball)

			source, err = filepath.EvalSymlinks(source)
			if err != nil {
				return log.Failed(err)
			}

			err = os.Rename(source, dst)
			if err != nil {
				return log.Failed(err)
			}
		}
	}

	args := docker.ContainerCreateArgs{
		Mounts: mounts,
		Image:  naming.Image(),
		Name:   naming.Container(),
		User:   fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
	}
	err = docker.ContainerCreate(args)
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

// Start function commands Docker Engine to start container.
func Start() error {
	log.Info("Starting container")

	isContainerStarted, err := docker.IsContainerStarted(naming.Container())
	if err != nil {
		return log.Failed(err)
	}
	if isContainerStarted {
		return log.Skipped()
	}

	err = docker.ContainerStart(naming.Container())
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

func Depends() error {
	log.Info("Installing dependencies")
	log.Drop()

	args := []docker.ContainerExecArgs{
		{
			Name:    naming.Container(),
			Cmd:     "rm -f a.list",
			AsRoot:  true,
			WorkDir: "/etc/apt/sources.list.d",
		}, {
			Name:    naming.Container(),
			Cmd:     "echo deb [trusted=yes] file://" + docker.ContainerArchiveDir + " ./ > a.list",
			AsRoot:  true,
			WorkDir: "/etc/apt/sources.list.d",
		}, {
			Name:    naming.Container(),
			Cmd:     "dpkg-scanpackages -m . > Packages",
			AsRoot:  true,
			WorkDir: docker.ContainerArchiveDir,
		}, {
			Name:    naming.Container(),
			Cmd:     "apt-get update",
			AsRoot:  true,
			Network: true,
			Skip:    NoUpdate,
		}, {
			Name:    naming.Container(),
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
		err := docker.ContainerExec(arg)
		if err != nil {
			return log.Failed(err)
		}
	}

	return log.Done()
}

// Package function first disables network in container,
// then executes "dpkg-buildpackage" and at the end,
// enables network back.
func Package() error {
	log.Info("Packaging software")
	log.Drop()

	args := docker.ContainerExecArgs{
		Name:    naming.Container(),
		Cmd:     "dpkg-buildpackage" + " " + DpkgFlags,
		Network: WithNetwork,
	}
	err := docker.ContainerExec(args)
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

// Test function executes "debi", "debc" and "lintian" in container.
func Test() error {
	log.Info("Testing package")
	log.Drop()

	args := []docker.ContainerExecArgs{
		{
			Name:    naming.Container(),
			Cmd:     "debi --with-depends",
			Network: true,
			AsRoot:  true,
		}, {
			Name: naming.Container(),
			Cmd:  "debc",
		}, {
			Name: naming.Container(),
			Cmd:  "lintian" + " " + LintianFlags,
		},
	}

	for _, arg := range args {
		err := docker.ContainerExec(arg)
		if err != nil {
			return log.Failed(err)
		}
	}

	return log.Done()
}

// Archive function moves successful build to archive by overwriting.
func Archive() error {
	log.Info("Archiving build")
	log.Drop()

	// Make needed directories
	err := os.MkdirAll(naming.ArchiveVersionDir(), os.ModePerm)
	if err != nil {
		return log.Failed(err)
	}

	// Read files in build directory
	files, err := ioutil.ReadDir(naming.BuildDir())
	if err != nil {
		return log.Failed(err)
	}

	for _, file := range files {
		// We don't need directories, only files
		if file.IsDir() {
			continue
		}

		log.ExtraInfo(file.Name())

		sourcePath := filepath.Join(naming.BuildDir(), file.Name())
		targetPath := filepath.Join(naming.ArchiveVersionDir(), file.Name())

		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			return log.Failed(err)
		}

		sourceBytes, err := ioutil.ReadAll(sourceFile)
		if err != nil {
			return log.Failed(err)
		}

		// Check if target file already exists
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

			// Compare checksums of source and target files
			//
			// if equal then simply skip copying this file
			if targetChecksum == sourceChecksum {
				_ = log.Skipped()
				continue
			}
		}

		// Target file doesn't exist or checksums mismatched
		err = ioutil.WriteFile(targetPath, sourceBytes, os.ModePerm)
		if err != nil {
			return log.Failed(err)
		}

		err = sourceFile.Close()
		if err != nil {
			return log.Failed(err)
		}

		_ = log.Done()
	}

	return log.None()
}

// Stop function commands Docker Engine to stop container.
func Stop() error {
	log.Info("Stopping container")

	isContainerStopped, err := docker.IsContainerStopped(naming.Container())
	if err != nil {
		return log.Failed(err)
	}
	if isContainerStopped {
		return log.Skipped()
	}

	err = docker.ContainerStop(naming.Container())
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

// Remove function commands Docker Engine to remove container.
func Remove() error {
	log.Info("Removing container")

	isContainerCreated, err := docker.IsContainerCreated(naming.Container())
	if err != nil {
		return log.Failed(err)
	}
	if !isContainerCreated {
		return log.Skipped()
	}

	err = docker.ContainerRemove(naming.Container())
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

// ShellOptional function interactively executes bash shell in container.
func ShellOptional() error {
	log.Info("Launching shell")
	log.Drop()

	args := docker.ContainerExecArgs{
		Interactive: true,
		AsRoot:      true,
		Name:        naming.Container(),
	}
	err := docker.ContainerExec(args)
	if err != nil {
		return log.Failed(err)
	}

	return log.None()
}

// Check function evaluates if package has been already built and
// is in archive, if it is, then it exits with 0 code.
func CheckOptional() error {
	log.Info("Checking archive")

	minFiles := 3
	foundFiles := 0
	err := walk.Walk(naming.ArchiveVersionDir(), 1, func(node *walk.Node) bool {
		foundFiles++
		return false
	})

	if err != nil || foundFiles < minFiles {
		return log.Custom("not built")
	}

	return log.Custom("already built")
}
