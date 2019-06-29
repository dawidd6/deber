package steps

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/dockerfile"
	"github.com/dawidd6/deber/pkg/dockerhub"
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/util"
	"github.com/dawidd6/deber/pkg/walk"
	"github.com/docker/docker/api/types/mount"
	"io/ioutil"
	"os"
	"path/filepath"
	"pault.ag/go/debian/changelog"
	"strings"
)

type Stepping struct {
	Docker *docker.Docker
	Log    *logger.Logger
	Naming *naming.Naming
	Debian *changelog.ChangelogEntry
}

// Build function determines parent image name by querying DockerHub API
// for available "debian" and "ubuntu" tags and confronting them with
// debian/changelog's target distribution.
//
// At last it commands Docker Engine to build image.
func (s *Stepping) Build(noRebuild bool) error {
	s.Log.Info("Building image")

	isImageBuilt, err := s.Docker.IsImageBuilt(s.Naming.Image())
	if err != nil {
		return err
	}
	if isImageBuilt {
		isImageOld, err := s.Docker.IsImageOld(s.Naming.Image())
		if err != nil {
			return err
		}
		if !isImageOld {
			return nil
		} else if noRebuild {
			return nil
		}
	}

	tag := strings.Split(s.Naming.Image(), ":")[1]
	repos := []string{"debian", "ubuntu"}
	repo, err := dockerhub.MatchRepo(repos, tag)
	if err != nil {
		return err
	}

	file, err := dockerfile.Parse(repo, tag)
	if err != nil {
		return err
	}

	args := docker.ImageBuildArgs{
		Name:       s.Naming.Image(),
		Dockerfile: file,
	}
	err = s.Docker.ImageBuild(args)
	if err != nil {
		return err
	}

	return nil
}

// Create function commands Docker Engine to create container.
//
// Also makes directories on host and moves tarball if needed.
func (s *Stepping) Create(extraPackages []string) error {
	s.Log.Info("Creating container")

	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: s.Naming.SourceDir(),
			Target: docker.ContainerSourceDir,
		}, {
			Type:   mount.TypeBind,
			Source: s.Naming.BuildDir(),
			Target: docker.ContainerBuildDir,
		}, {
			Type:   mount.TypeBind,
			Source: s.Naming.CacheDir(),
			Target: docker.ContainerCacheDir,
		},
	}

	// Handle extra packages mounting
	for _, pkg := range extraPackages {
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

	isContainerCreated, err := s.Docker.IsContainerCreated(s.Naming.Container())
	if err != nil {
		return err
	}
	if isContainerCreated {
		oldMounts, err := s.Docker.ContainerMounts(s.Naming.Container())
		if err != nil {
			return err
		}

		// Compare old mounts with new ones,
		// if not equal, then recreate container
		if util.CompareMounts(oldMounts, mounts) {
			return nil
		}

		err = s.Docker.ContainerStop(s.Naming.Container())
		if err != nil {
			return err
		}

		err = s.Docker.ContainerRemove(s.Naming.Container())
		if err != nil {
			return err
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
			return err
		}
	}

	// Find tarball if package is not native
	if !s.Debian.Version.IsNative() {
		tarball := fmt.Sprintf("%s_%s.orig.tar", s.Debian.Source, s.Debian.Version.Version)
		found := false

		// First check source parent directory
		files, err := ioutil.ReadDir(s.Naming.SourceParentDir())
		if err != nil {
			return err
		}

		for _, file := range files {
			if strings.HasPrefix(file.Name(), tarball) {
				tarball = file.Name()
				found = true
				break
			}
		}

		// If tarball is present there, then move it to build directory
		if found {
			source := filepath.Join(s.Naming.SourceParentDir(), tarball)
			dst := filepath.Join(s.Naming.BuildDir(), tarball)

			source, err = filepath.EvalSymlinks(source)
			if err != nil {
				return err
			}
			// TODO remove all orig tarballs in build dir before moving a new one
			err = os.Rename(source, dst)
			if err != nil {
				return err
			}
		} else {
			files, err := ioutil.ReadDir(s.Naming.BuildDir())
			if err != nil {
				return err
			}

			for _, file := range files {
				if strings.HasPrefix(file.Name(), tarball) {
					tarball = file.Name()
					found = true
					break
				}
			}

			if !found {
				return errors.New("upstream tarball not found")
			}
		}
	}

	args := docker.ContainerCreateArgs{
		Mounts: mounts,
		Image:  s.Naming.Image(),
		Name:   s.Naming.Container(),
		User:   fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
	}
	err = s.Docker.ContainerCreate(args)
	if err != nil {
		return err
	}

	return nil
}

// Start function commands Docker Engine to start container.
func (s *Stepping) Start() error {
	s.Log.Info("Starting container")

	isContainerStarted, err := s.Docker.IsContainerStarted(s.Naming.Container())
	if err != nil {
		return err
	}
	if isContainerStarted {
		return nil
	}

	err = s.Docker.ContainerStart(s.Naming.Container())
	if err != nil {
		return err
	}

	return nil
}

func (s *Stepping) Depends(extraPackages []string, noUpdate bool) error {
	s.Log.Info("Installing dependencies")

	args := []docker.ContainerExecArgs{
		{
			Name:    s.Naming.Container(),
			Cmd:     "rm -f a.list",
			AsRoot:  true,
			WorkDir: "/etc/apt/sources.list.d",
		}, {
			Name:    s.Naming.Container(),
			Cmd:     "echo deb [trusted=yes] file://" + docker.ContainerArchiveDir + " ./ > a.list",
			AsRoot:  true,
			WorkDir: "/etc/apt/sources.list.d",
		}, {
			Name:    s.Naming.Container(),
			Cmd:     "dpkg-scanpackages -m . > Packages",
			AsRoot:  true,
			WorkDir: docker.ContainerArchiveDir,
		}, {
			Name:    s.Naming.Container(),
			Cmd:     "apt-get update",
			AsRoot:  true,
			Network: true,
			Skip:    noUpdate,
		}, {
			Name:    s.Naming.Container(),
			Cmd:     "apt-get build-dep ./",
			Network: true,
			AsRoot:  true,
		},
	}

	if extraPackages == nil {
		args[1].Skip = true
		args[2].Skip = true
	}

	for _, arg := range args {
		err := s.Docker.ContainerExec(arg)
		if err != nil {
			return err
		}
	}

	return nil
}

// Package function first disables network in container,
// then executes "dpkg-buildpackage" and at the end,
// enables network back.
func (s *Stepping) Package(dpkgFlags string, withNetwork bool) error {
	s.Log.Info("Packaging software")

	args := docker.ContainerExecArgs{
		Name:    s.Naming.Container(),
		Cmd:     "dpkg-buildpackage" + " " + dpkgFlags,
		Network: withNetwork,
	}
	err := s.Docker.ContainerExec(args)
	if err != nil {
		return err
	}

	return nil
}

// Test function executes "debi", "debc" and "lintian" in container.
func (s *Stepping) Test(lintianFlags string) error {
	s.Log.Info("Testing package")

	args := []docker.ContainerExecArgs{
		{
			Name:    s.Naming.Container(),
			Cmd:     "debi --with-depends",
			Network: true,
			AsRoot:  true,
		}, {
			Name: s.Naming.Container(),
			Cmd:  "debc",
		}, {
			Name: s.Naming.Container(),
			Cmd:  "lintian" + " " + lintianFlags,
		},
	}

	for _, arg := range args {
		err := s.Docker.ContainerExec(arg)
		if err != nil {
			return err
		}
	}

	return nil
}

// Archive function moves successful build to archive by overwriting.
func (s *Stepping) Archive() error {
	s.Log.Info("Archiving build")

	// Make needed directories
	err := os.MkdirAll(s.Naming.ArchiveVersionDir(), os.ModePerm)
	if err != nil {
		return err
	}

	// Read files in build directory
	files, err := ioutil.ReadDir(s.Naming.BuildDir())
	if err != nil {
		return err
	}

	for _, file := range files {
		// We don't need directories, only files
		if file.IsDir() {
			continue
		}

		sourcePath := filepath.Join(s.Naming.BuildDir(), file.Name())
		targetPath := filepath.Join(s.Naming.ArchiveVersionDir(), file.Name())

		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			return err
		}

		sourceBytes, err := ioutil.ReadAll(sourceFile)
		if err != nil {
			return err
		}

		// Check if target file already exists
		targetStat, _ := os.Stat(targetPath)
		if targetStat != nil {
			targetFile, err := os.Open(targetPath)
			if err != nil {
				return err
			}

			targetBytes, err := ioutil.ReadAll(targetFile)
			if err != nil {
				return err
			}

			sourceChecksum := md5.Sum(sourceBytes)
			targetChecksum := md5.Sum(targetBytes)

			// Compare checksums of source and target files
			//
			// if equal then simply skip copying this file
			if targetChecksum == sourceChecksum {
				continue
			}
		}

		// Target file doesn't exist or checksums mismatched
		err = ioutil.WriteFile(targetPath, sourceBytes, os.ModePerm)
		if err != nil {
			return err
		}

		err = sourceFile.Close()
		if err != nil {
			return err
		}

		s.Log.InfoExtra(file.Name())
	}

	return nil
}

// Stop function commands Docker Engine to stop container.
func (s *Stepping) Stop() error {
	s.Log.Info("Stopping container")

	isContainerStopped, err := s.Docker.IsContainerStopped(s.Naming.Container())
	if err != nil {
		return err
	}
	if isContainerStopped {
		return nil
	}

	err = s.Docker.ContainerStop(s.Naming.Container())
	if err != nil {
		return err
	}

	return nil
}

// Remove function commands Docker Engine to remove container.
func (s *Stepping) Remove() error {
	s.Log.Info("Removing container")

	isContainerCreated, err := s.Docker.IsContainerCreated(s.Naming.Container())
	if err != nil {
		return err
	}
	if !isContainerCreated {
		return nil
	}

	err = s.Docker.ContainerRemove(s.Naming.Container())
	if err != nil {
		return err
	}

	return nil
}

// ShellOptional function interactively executes bash shell in container.
func (s *Stepping) ShellOptional() error {
	s.Log.Info("Launching shell")

	args := docker.ContainerExecArgs{
		Interactive: true,
		AsRoot:      true,
		Network:     true,
		Name:        s.Naming.Container(),
	}
	err := s.Docker.ContainerExec(args)
	if err != nil {
		return err
	}

	return nil
}

// Check function evaluates if package has been already built and
// is in archive, if it is, then it exits with 0 code.
func (s *Stepping) CheckOptional() error {
	s.Log.Info("Checking archive")

	minFiles := 3
	foundFiles := 0
	err := walk.Walk(s.Naming.ArchiveVersionDir(), 1, func(node *walk.Node) bool {
		foundFiles++
		return false
	})
	if err != nil {
		return err
	}

	if foundFiles < minFiles {
		s.Log.InfoExtra("not built")
		return nil
	} else {
		s.Log.InfoExtra("already built")
		return nil
	}
}
