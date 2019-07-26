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
	"github.com/dawidd6/deber/pkg/utils"
	"github.com/docker/docker/api/types/mount"
	"io/ioutil"
	"os"
	"path/filepath"
	"pault.ag/go/debian/changelog"
	"strings"
	"time"
)

// Build function determines parent image name by querying DockerHub API
// for available "debian" and "ubuntu" tags and confronting them with
// debian/changelog's target distribution.
//
// At last it commands Docker Engine to build image.
func Build(dock *docker.Docker, n *naming.Naming, maxAge time.Duration) error {
	log.Info("Building image")

	isImageBuilt, err := dock.IsImageBuilt(n.Image())
	if err != nil {
		return log.Failed(err)
	}
	if isImageBuilt {
		age, err := dock.ImageAge(n.Image())
		if err != nil {
			return log.Failed(err)
		}

		if age.Hours() < maxAge.Hours() {
			return log.Skipped()
		}
	}

	tag := strings.Split(n.Image(), ":")[1]
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
		Name:       n.Image(),
		Dockerfile: file,
	}
	err = dock.ImageBuild(args)
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

// Create function commands Docker Engine to create container.
//
// Also makes directories on host and moves tarball if needed.
func Create(dock *docker.Docker, n *naming.Naming, extraPackages []string) error {
	log.Info("Creating container")

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

	// Handle extra packages mounting
	for _, pkg := range extraPackages {
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

	isContainerCreated, err := dock.IsContainerCreated(n.Container())
	if err != nil {
		return log.Failed(err)
	}
	if isContainerCreated {
		oldMounts, err := dock.ContainerMounts(n.Container())
		if err != nil {
			return log.Failed(err)
		}

		// Compare old mounts with new ones,
		// if not equal, then recreate container
		if utils.CompareMounts(oldMounts, mounts) {
			return log.Skipped()
		}

		err = dock.ContainerStop(n.Container())
		if err != nil {
			return log.Failed(err)
		}

		err = dock.ContainerRemove(n.Container())
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

	user := fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())
	args := docker.ContainerCreateArgs{
		Mounts: mounts,
		Image:  n.Image(),
		Name:   n.Container(),
		User:   user,
	}
	err = dock.ContainerCreate(args)
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

// Start function commands Docker Engine to start container.
func Start(dock *docker.Docker, n *naming.Naming) error {
	log.Info("Starting container")

	isContainerStarted, err := dock.IsContainerStarted(n.Container())
	if err != nil {
		return log.Failed(err)
	}
	if isContainerStarted {
		return log.Skipped()
	}

	err = dock.ContainerStart(n.Container())
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

func Tarball(n *naming.Naming, deb *changelog.ChangelogEntry) error {
	log.Info("Finding tarballs")

	if deb.Version.IsNative() {
		return log.Skipped()
	}

	tarball := fmt.Sprintf("%s_%s.orig.tar", deb.Source, deb.Version.Version)

	sourceTarballs := make([]string, 0)
	sourceFiles, err := ioutil.ReadDir(n.SourceParentDir())
	if err != nil {
		return log.Failed(err)
	}

	buildTarballs := make([]string, 0)
	buildFiles, err := ioutil.ReadDir(n.BuildDir())
	if err != nil {
		return log.Failed(err)
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
		return log.Failed(errors.New("multiple tarballs found in build directory"))
	}

	if len(sourceTarballs) > 1 {
		return log.Failed(errors.New("multiple tarballs found in parent source directory"))
	}

	if len(sourceTarballs) < 1 && len(buildTarballs) < 1 {
		return log.Failed(errors.New("upstream tarball not found"))
	}

	if len(sourceTarballs) == 1 {
		if len(buildTarballs) == 1 {
			f := filepath.Join(n.BuildDir(), buildTarballs[0])
			err = os.Remove(f)
			if err != nil {
				return log.Failed(err)
			}
		}

		src := filepath.Join(n.SourceParentDir(), sourceTarballs[0])
		dst := filepath.Join(n.BuildDir(), sourceTarballs[0])

		src, err = filepath.EvalSymlinks(src)
		if err != nil {
			return log.Failed(err)
		}

		err = os.Rename(src, dst)
		if err != nil {
			return log.Failed(err)
		}
	} else {
		return log.Skipped()
	}

	return log.Done()
}

func Depends(dock *docker.Docker, n *naming.Naming, extraPackages []string) error {
	log.Info("Installing dependencies")
	log.Drop()

	args := []docker.ContainerExecArgs{
		{
			Name:    n.Container(),
			Cmd:     "rm -f a.list",
			AsRoot:  true,
			WorkDir: "/etc/apt/sources.list.d",
		}, {
			Name:    n.Container(),
			Cmd:     "echo deb [trusted=yes] file://" + docker.ContainerArchiveDir + " ./ > a.list",
			AsRoot:  true,
			WorkDir: "/etc/apt/sources.list.d",
		}, {
			Name:    n.Container(),
			Cmd:     "dpkg-scanpackages -m . > Packages",
			AsRoot:  true,
			WorkDir: docker.ContainerArchiveDir,
		}, {
			Name:    n.Container(),
			Cmd:     "apt-get update",
			AsRoot:  true,
			Network: true,
		}, {
			Name:    n.Container(),
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
func Package(dock *docker.Docker, n *naming.Naming, dpkgFlags string, withNetwork bool) error {
	log.Info("Packaging software")
	log.Drop()

	args := docker.ContainerExecArgs{
		Name:    n.Container(),
		Cmd:     "dpkg-buildpackage" + " " + dpkgFlags,
		Network: withNetwork,
	}
	err := dock.ContainerExec(args)
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

// Test function executes "debi", "debc" and "lintian" in container.
func Test(dock *docker.Docker, n *naming.Naming, lintianFlags string) error {
	log.Info("Testing package")
	log.Drop()

	args := []docker.ContainerExecArgs{
		{
			Name:    n.Container(),
			Cmd:     "debi --with-depends",
			Network: true,
			AsRoot:  true,
		}, {
			Name: n.Container(),
			Cmd:  "debc",
		}, {
			Name: n.Container(),
			Cmd:  "lintian" + " " + lintianFlags,
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
func Archive(n *naming.Naming) error {
	log.Info("Archiving build")

	// Make needed directories
	err := os.MkdirAll(n.ArchiveVersionDir(), os.ModePerm)
	if err != nil {
		return log.Failed(err)
	}

	// Read files in build directory
	files, err := ioutil.ReadDir(n.BuildDir())
	if err != nil {
		return log.Failed(err)
	}

	log.Drop()

	for _, f := range files {
		// We don't need directories, only files
		if f.IsDir() {
			continue
		}

		log.ExtraInfo(f.Name())

		sourcePath := filepath.Join(n.BuildDir(), f.Name())
		targetPath := filepath.Join(n.ArchiveVersionDir(), f.Name())

		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			return log.Failed(err)
		}

		sourceBytes, err := ioutil.ReadAll(sourceFile)
		if err != nil {
			return log.Failed(err)
		}

		sourceStat, err := sourceFile.Stat()
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
		err = ioutil.WriteFile(targetPath, sourceBytes, sourceStat.Mode())
		if err != nil {
			return log.Failed(err)
		}

		err = sourceFile.Close()
		if err != nil {
			return log.Failed(err)
		}

		_ = log.Done()
	}

	log.Drop()
	return log.Done()
}

// Stop function commands Docker Engine to stop container.
func Stop(dock *docker.Docker, n *naming.Naming) error {
	log.Info("Stopping container")

	isContainerStopped, err := dock.IsContainerStopped(n.Container())
	if err != nil {
		return log.Failed(err)
	}
	if isContainerStopped {
		return log.Skipped()
	}

	err = dock.ContainerStop(n.Container())
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

// Remove function commands Docker Engine to remove container.
func Remove(dock *docker.Docker, n *naming.Naming) error {
	log.Info("Removing container")

	isContainerCreated, err := dock.IsContainerCreated(n.Container())
	if err != nil {
		return log.Failed(err)
	}
	if !isContainerCreated {
		return log.Skipped()
	}

	err = dock.ContainerRemove(n.Container())
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

// ShellOptional function interactively executes bash shell in container.
func ShellOptional(dock *docker.Docker, n *naming.Naming) error {
	log.Info("Launching shell")
	log.Drop()

	args := docker.ContainerExecArgs{
		Interactive: true,
		AsRoot:      true,
		Network:     true,
		Name:        n.Container(),
	}
	err := dock.ContainerExec(args)
	if err != nil {
		return log.Failed(err)
	}

	return log.Done()
}

// Check function evaluates if package has been already built and
// is in archive, if it is, then it exits with 0 code.
func CheckOptional(n *naming.Naming) error {
	log.Info("Checking archive")
	log.Drop()

	minFiles := 3
	foundFiles := 0
	err := utils.Walk(n.ArchiveVersionDir(), 1, func(file *utils.File) bool {
		foundFiles++
		return false
	})

	if err == nil && foundFiles > minFiles {
		return log.Failed(err)
	}

	return log.Done()
}
