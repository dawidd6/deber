package stepping

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/docker/file"
	"github.com/dawidd6/deber/pkg/docker/hub"
	"github.com/dawidd6/deber/pkg/logger"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/dawidd6/deber/pkg/utils"
	"github.com/docker/docker/api/types/mount"
	"io/ioutil"
	"os"
	"path/filepath"
	"pault.ag/go/debian/changelog"
	"strings"
	"time"
)

type Stepping struct {
	Docker *docker.Docker
	Log    *logger.Logger
	Naming *naming.Naming
	Debian *changelog.ChangelogEntry

	DpkgFlags          string
	LintianFlags       string
	PackageWithNetwork bool
	RebuildImageIfOld  bool
	MaxImageAge        time.Duration
	ExtraPackages      []string
}

// Build function determines parent image name by querying DockerHub API
// for available "debian" and "ubuntu" tags and confronting them with
// debian/changelog's target distribution.
//
// At last it commands Docker Engine to build image.
func (s *Stepping) Build() error {
	s.Log.Info("Building image")

	isImageBuilt, err := s.Docker.IsImageBuilt(s.Naming.Image())
	if err != nil {
		return s.Log.Failed(err)
	}
	if isImageBuilt {
		age, err := s.Docker.ImageAge(s.Naming.Image())
		if err != nil {
			return s.Log.Failed(err)
		}

		if age.Hours() < s.MaxImageAge.Hours() {
			return s.Log.Skipped()
		} else if !s.RebuildImageIfOld {
			return s.Log.Skipped()
		}
	}

	tag := strings.Split(s.Naming.Image(), ":")[1]
	repos := []string{"debian", "ubuntu"}
	repo, err := hub.MatchRepo(repos, tag)
	if err != nil {
		return s.Log.Failed(err)
	}

	dockerfile, err := file.Parse(repo, tag)
	if err != nil {
		return s.Log.Failed(err)
	}

	s.Log.Drop()

	args := docker.ImageBuildArgs{
		Name:       s.Naming.Image(),
		Dockerfile: dockerfile,
	}
	err = s.Docker.ImageBuild(args)
	if err != nil {
		return s.Log.Failed(err)
	}

	return s.Log.Done()
}

// Create function commands Docker Engine to create container.
//
// Also makes directories on host and moves tarball if needed.
func (s *Stepping) Create() error {
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
	for _, pkg := range s.ExtraPackages {
		source, err := filepath.Abs(pkg)
		if err != nil {
			return s.Log.Failed(err)
		}

		info, err := os.Stat(source)
		if info == nil {
			return s.Log.Failed(err)
		}
		if !info.IsDir() && !strings.HasSuffix(source, ".deb") {
			return s.Log.Failed(errors.New("please specify a directory or .deb file"))
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
		return s.Log.Failed(err)
	}
	if isContainerCreated {
		oldMounts, err := s.Docker.ContainerMounts(s.Naming.Container())
		if err != nil {
			return s.Log.Failed(err)
		}

		// Compare old mounts with new ones,
		// if not equal, then recreate container
		if utils.CompareMounts(oldMounts, mounts) {
			return s.Log.Skipped()
		}

		err = s.Docker.ContainerStop(s.Naming.Container())
		if err != nil {
			return s.Log.Failed(err)
		}

		err = s.Docker.ContainerRemove(s.Naming.Container())
		if err != nil {
			return s.Log.Failed(err)
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
			return s.Log.Failed(err)
		}
	}

	user := fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())
	args := docker.ContainerCreateArgs{
		Mounts: mounts,
		Image:  s.Naming.Image(),
		Name:   s.Naming.Container(),
		User:   user,
	}
	err = s.Docker.ContainerCreate(args)
	if err != nil {
		return s.Log.Failed(err)
	}

	return s.Log.Done()
}

// Start function commands Docker Engine to start container.
func (s *Stepping) Start() error {
	s.Log.Info("Starting container")

	isContainerStarted, err := s.Docker.IsContainerStarted(s.Naming.Container())
	if err != nil {
		return s.Log.Failed(err)
	}
	if isContainerStarted {
		return s.Log.Skipped()
	}

	err = s.Docker.ContainerStart(s.Naming.Container())
	if err != nil {
		return s.Log.Failed(err)
	}

	return s.Log.Done()
}

func (s *Stepping) Tarball() error {
	s.Log.Info("Finding tarballs")

	if s.Debian.Version.IsNative() {
		return s.Log.Skipped()
	}

	tarball := fmt.Sprintf("%s_%s.orig.tar", s.Debian.Source, s.Debian.Version.Version)

	sourceTarballs := make([]string, 0)
	sourceFiles, err := ioutil.ReadDir(s.Naming.SourceParentDir())
	if err != nil {
		return s.Log.Failed(err)
	}

	buildTarballs := make([]string, 0)
	buildFiles, err := ioutil.ReadDir(s.Naming.BuildDir())
	if err != nil {
		return s.Log.Failed(err)
	}

	for _, f := range sourceFiles {
		if strings.HasPrefix(f.Name(), tarball) {
			sourceTarballs = append(sourceTarballs, f.Name())
		}
	}

	for _, f := range buildFiles {
		if strings.HasPrefix(f.Name(), tarball) {
			buildTarballs = append(buildTarballs, f.Name())
		}
	}

	if len(buildTarballs) > 1 {
		return s.Log.Failed(errors.New("multiple tarballs found in build directory"))
	}

	if len(sourceTarballs) > 1 {
		return s.Log.Failed(errors.New("multiple tarballs found in parent source directory"))
	}

	if len(sourceTarballs) < 1 && len(buildTarballs) < 1 {
		return s.Log.Failed(errors.New("upstream tarball not found"))
	}

	if len(sourceTarballs) == 1 {
		if len(buildTarballs) == 1 {
			f := filepath.Join(s.Naming.BuildDir(), buildTarballs[0])
			err = os.Remove(f)
			if err != nil {
				return s.Log.Failed(err)
			}
		}

		src := filepath.Join(s.Naming.SourceParentDir(), sourceTarballs[0])
		dst := filepath.Join(s.Naming.BuildDir(), sourceTarballs[0])

		src, err = filepath.EvalSymlinks(src)
		if err != nil {
			return s.Log.Failed(err)
		}

		err = os.Rename(src, dst)
		if err != nil {
			return s.Log.Failed(err)
		}
	} else {
		return s.Log.Skipped()
	}

	return s.Log.Done()
}

func (s *Stepping) Depends() error {
	s.Log.Info("Installing dependencies")
	s.Log.Drop()

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
		}, {
			Name:    s.Naming.Container(),
			Cmd:     "apt-get build-dep ./",
			Network: true,
			AsRoot:  true,
		},
	}

	if s.ExtraPackages == nil {
		args[1].Skip = true
		args[2].Skip = true
	}

	for _, arg := range args {
		err := s.Docker.ContainerExec(arg)
		if err != nil {
			return s.Log.Failed(err)
		}
	}

	return s.Log.Done()
}

// Package function first disables network in container,
// then executes "dpkg-buildpackage" and at the end,
// enables network back.
func (s *Stepping) Package() error {
	s.Log.Info("Packaging software")
	s.Log.Drop()

	args := docker.ContainerExecArgs{
		Name:    s.Naming.Container(),
		Cmd:     "dpkg-buildpackage" + " " + s.DpkgFlags,
		Network: s.PackageWithNetwork,
	}
	err := s.Docker.ContainerExec(args)
	if err != nil {
		return s.Log.Failed(err)
	}

	return s.Log.Done()
}

// Test function executes "debi", "debc" and "lintian" in container.
func (s *Stepping) Test() error {
	s.Log.Info("Testing package")
	s.Log.Drop()

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
			Cmd:  "lintian" + " " + s.LintianFlags,
		},
	}

	for _, arg := range args {
		err := s.Docker.ContainerExec(arg)
		if err != nil {
			return s.Log.Failed(err)
		}
	}

	return s.Log.Done()
}

// Archive function moves successful build to archive by overwriting.
func (s *Stepping) Archive() error {
	s.Log.Info("Archiving build")

	// Make needed directories
	err := os.MkdirAll(s.Naming.ArchiveVersionDir(), os.ModePerm)
	if err != nil {
		return s.Log.Failed(err)
	}

	// Read files in build directory
	files, err := ioutil.ReadDir(s.Naming.BuildDir())
	if err != nil {
		return s.Log.Failed(err)
	}

	s.Log.Drop()

	for _, f := range files {
		// We don't need directories, only files
		if f.IsDir() {
			continue
		}

		s.Log.ExtraInfo(f.Name())

		sourcePath := filepath.Join(s.Naming.BuildDir(), f.Name())
		targetPath := filepath.Join(s.Naming.ArchiveVersionDir(), f.Name())

		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			return s.Log.Failed(err)
		}

		sourceBytes, err := ioutil.ReadAll(sourceFile)
		if err != nil {
			return s.Log.Failed(err)
		}

		sourceStat, err := sourceFile.Stat()
		if err != nil {
			return s.Log.Failed(err)
		}

		// Check if target file already exists
		targetStat, _ := os.Stat(targetPath)
		if targetStat != nil {
			targetFile, err := os.Open(targetPath)
			if err != nil {
				return s.Log.Failed(err)
			}

			targetBytes, err := ioutil.ReadAll(targetFile)
			if err != nil {
				return s.Log.Failed(err)
			}

			sourceChecksum := md5.Sum(sourceBytes)
			targetChecksum := md5.Sum(targetBytes)

			// Compare checksums of source and target files
			//
			// if equal then simply skip copying this file
			if targetChecksum == sourceChecksum {
				_ = s.Log.Skipped()
				continue
			}
		}

		// Target file doesn't exist or checksums mismatched
		err = ioutil.WriteFile(targetPath, sourceBytes, sourceStat.Mode())
		if err != nil {
			return s.Log.Failed(err)
		}

		err = sourceFile.Close()
		if err != nil {
			return s.Log.Failed(err)
		}

		_ = s.Log.Done()
	}

	s.Log.Drop()
	return s.Log.Done()
}

// Stop function commands Docker Engine to stop container.
func (s *Stepping) Stop() error {
	s.Log.Info("Stopping container")

	isContainerStopped, err := s.Docker.IsContainerStopped(s.Naming.Container())
	if err != nil {
		return s.Log.Failed(err)
	}
	if isContainerStopped {
		return s.Log.Skipped()
	}

	err = s.Docker.ContainerStop(s.Naming.Container())
	if err != nil {
		return s.Log.Failed(err)
	}

	return s.Log.Done()
}

// Remove function commands Docker Engine to remove container.
func (s *Stepping) Remove() error {
	s.Log.Info("Removing container")

	isContainerCreated, err := s.Docker.IsContainerCreated(s.Naming.Container())
	if err != nil {
		return s.Log.Failed(err)
	}
	if !isContainerCreated {
		return s.Log.Skipped()
	}

	err = s.Docker.ContainerRemove(s.Naming.Container())
	if err != nil {
		return s.Log.Failed(err)
	}

	return s.Log.Done()
}

// ShellOptional function interactively executes bash shell in container.
func (s *Stepping) ShellOptional() error {
	s.Log.Info("Launching shell")
	s.Log.Drop()

	args := docker.ContainerExecArgs{
		Interactive: true,
		AsRoot:      true,
		Network:     true,
		Name:        s.Naming.Container(),
	}
	err := s.Docker.ContainerExec(args)
	if err != nil {
		return s.Log.Failed(err)
	}

	return s.Log.Done()
}

// Check function evaluates if package has been already built and
// is in archive, if it is, then it exits with 0 code.
func (s *Stepping) CheckOptional() error {
	s.Log.Info("Checking archive")
	s.Log.Drop()

	minFiles := 3
	foundFiles := 0
	err := utils.Walk(s.Naming.ArchiveVersionDir(), 1, func(file *utils.File) bool {
		foundFiles++
		return false
	})

	if err == nil && foundFiles > minFiles {
		return s.Log.Failed(err)
	}

	return s.Log.Done()
}
