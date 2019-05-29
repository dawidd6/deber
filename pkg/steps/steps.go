package steps

import (
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/docker/docker/api/types/mount"
	"os"
	"path/filepath"
	"strings"
)

// Build function determines parent image name by querying DockerHub API
// for available "debian" and "ubuntu" tags and confronting them with
// debian/changelog's target distribution.
//
// At last it commands Docker Engine to build image.
func Build(dock *docker.Docker, args BuildArgs) error {
	log.Info("Building image")

	isImageBuilt, err := dock.IsImageBuilt(args.Image.Name())
	if err != nil {
		return log.FailE(err)
	}
	if isImageBuilt && !args.IsRebuildNeeded {
		isImageOld, err := dock.IsImageOld(args.Image.Name())
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
			if tag.Name == args.Image.Tag() {
				from := fmt.Sprintf("%s:%s", repo, args.Image.Tag())

				log.Drop()

				args := docker.ImageBuildArgs{
					From: from,
					Name: args.Image.Name(),
				}
				err := dock.ImageBuild(args)
				if err != nil {
					return log.FailE(err)
				}

				return log.DoneE()
			}
		}
	}

	return log.FailE(errors.New("distribution image not found"))
}

// Create function commands Docker Engine to create and start container.
func Create(dock *docker.Docker, args CreateArgs) error {
	log.Info("Creating container")

	isContainerCreated, err := dock.IsContainerCreated(args.ContainerName)
	if err != nil {
		return log.FailE(err)
	}
	if isContainerCreated {
		return log.SkipE()
	}

	containerArgs := docker.ContainerCreateArgs{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: args.SourceDir,
				Target: docker.ContainerSourceDir,
			}, {
				Type:   mount.TypeBind,
				Source: args.BuildDir,
				Target: docker.ContainerBuildDir,
			}, {
				Type:   mount.TypeBind,
				Source: args.CacheDir,
				Target: docker.ContainerCacheDir,
			},
		},
		Image: args.ImageName,
		Name:  args.ContainerName,
		User:  fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
	}

	for _, mnt := range containerArgs.Mounts {
		err := os.MkdirAll(mnt.Source, os.ModePerm)
		if err != nil {
			return log.FailE(err)
		}
	}

	// TODO check
	for _, pkg := range args.ExtraPackages {
		source, err := filepath.Abs(pkg)
		if err != nil {
			return log.FailE(err)
		}

		info, err := os.Stat(source)
		if info == nil {
			return log.FailE(err)
		}
		if !info.IsDir() && !strings.HasSuffix(source, ".deb") {
			return log.FailE(errors.New("please specify a directory or .deb file"))
		}

		target := filepath.Join(docker.ContainerArchiveDir, filepath.Base(source))

		mnt := mount.Mount{
			Type:     mount.TypeBind,
			Source:   source,
			Target:   target,
			ReadOnly: true,
		}

		containerArgs.Mounts = append(containerArgs.Mounts, mnt)
	}

	err = dock.ContainerCreate(containerArgs)
	if err != nil {
		return log.FailE(err)
	}

	err = dock.ContainerStart(args.ContainerName)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func Depends(dock *docker.Docker, args DependsArgs) error {
	log.Info("Installing dependencies")

	log.Drop()

	containerNetworkArgs := docker.ContainerNetworkArgs{
		Name:      args.ContainerName,
		Connected: true,
	}

	err := dock.ContainerNetwork(containerNetworkArgs)
	if err != nil {
		return log.FailE(err)
	}

	commands := []string{
		"apt-get update",
		"apt-get build-dep " + docker.ContainerSourceDir,
	}

	if args.ExtraPackages != nil {
		commands = append(
			[]string{
				fmt.Sprintf(
					"echo deb [trusted=yes] file://%s %s > %s\n",
					docker.ContainerArchiveDir,
					"./",
					"/etc/apt/sources.list.d/a.list",
				),
				"dpkg-scanpackages -m . > Packages",
			},
			commands...,
		)
	}

	for _, cmd := range commands {
		containerArgs := docker.ContainerExecArgs{
			Name:    args.ContainerName,
			Cmd:     cmd,
			WorkDir: docker.ContainerArchiveDir,
			AsRoot:  true,
		}

		err := dock.ContainerExec(containerArgs)
		if err != nil {
			return log.FailE(err)
		}
	}

	return log.DoneE()
}

func Package(dock *docker.Docker, args PackageArgs) error {
	log.Info("Packaging software")

	log.Drop()

	if strings.Contains(args.PackageVersion, "-") {
		tarball := fmt.Sprintf(
			"%s_%s.orig.tar",
			args.PackageName,
			strings.Split(args.PackageVersion, "-")[0],
		)

		found := false
		path := filepath.Join(args.TarballSourceDir, tarball)

		for _, ext := range []string{".gz", ".xz", "bz2"} {
			info, _ := os.Stat(path + ext)
			if info != nil {
				tarball += ext
				source := filepath.Join(args.TarballSourceDir, tarball)
				target := filepath.Join(args.TarballTargetDir, tarball)

				source, err := filepath.EvalSymlinks(source)
				if err != nil {
					return log.FailE(err)
				}

				err = os.Rename(source, target)
				if err != nil {
					return log.FailE(err)
				}

				found = true
				break
			}
		}

		if !found {
			return log.FailE(errors.New("tarball not found"))
		}
	}

	containerNetworkArgs := docker.ContainerNetworkArgs{
		Name:      args.ContainerName,
		Connected: args.IsNetworkNeeded,
	}

	containerArgs := docker.ContainerExecArgs{
		Name: args.ContainerName,
		Cmd:  "dpkg-buildpackage" + " " + args.DpkgFlags,
	}

	commands := []string{
		"debc",
		"debi --with-depends",
		"lintian" + " " + args.LintianFlags,
	}

	err := dock.ContainerNetwork(containerNetworkArgs)
	if err != nil {
		return log.FailE(err)
	}

	err = dock.ContainerExec(containerArgs)
	if err != nil {
		return log.FailE(err)
	}

	if args.IsTestNeeded {
		for _, cmd := range commands {
			containerArgs := docker.ContainerExecArgs{
				Name:   args.ContainerName,
				Cmd:    cmd,
				AsRoot: true,
			}

			err := dock.ContainerExec(containerArgs)
			if err != nil {
				return log.FailE(err)
			}
		}
	}

	return log.DoneE()
}

// Archive function moves successful build to archive by overwriting.
func Archive(dock *docker.Docker, args ArchiveArgs) error {
	log.Info("Archiving build")

	info, _ := os.Stat(args.ArchivePackageDir)
	if info != nil {
		err := os.RemoveAll(args.ArchivePackageDir)
		if err != nil {
			return log.FailE(err)
		}
	}

	err := os.MkdirAll(filepath.Dir(args.ArchivePackageDir), os.ModePerm)
	if err != nil {
		return log.FailE(err)
	}

	err = os.Rename(args.BuildDir, args.ArchivePackageDir)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

// Remove function commands Docker Engine to stop and remove container.
func Remove(dock *docker.Docker, args RemoveArgs) error {
	log.Info("Removing container")

	isContainerCreated, err := dock.IsContainerCreated(args.ContainerName)
	if err != nil {
		return log.FailE(err)
	}
	if !isContainerCreated {
		return log.SkipE()
	}

	isContainerStopped, err := dock.IsContainerStopped(args.ContainerName)
	if err != nil {
		return log.FailE(err)
	}
	if !isContainerStopped {
		err = dock.ContainerStop(args.ContainerName)
		if err != nil {
			return log.FailE(err)
		}
	}

	err = dock.ContainerRemove(args.ContainerName)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

// Shell function interactively executes bash shell in container.
func Shell(dock *docker.Docker, args ShellArgs) error {
	log.Info("Launching shell")

	log.Drop()

	containerArgs := docker.ContainerExecArgs{
		Interactive: true,
		AsRoot:      true,
		Name:        args.ContainerName,
	}
	err := dock.ContainerExec(containerArgs)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}
