package steps

import (
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/log"
	"github.com/dawidd6/deber/pkg/naming"
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

type Options struct {
	*naming.Naming

	DpkgFlags     string
	LintianFlags  string
	Network       bool
	Rebuild       bool
	All           bool
	ExtraPackages []string
}

func determineRepo(dist string) (string, error) {
	repos := []string{"debian", "ubuntu"}

	for _, repo := range repos {
		tags, err := docker.GetTags(repo)
		if err != nil {
			return "", err
		}

		for _, tag := range tags {
			if tag.Name == dist {
				return repo, nil
			}
		}
	}

	return "", errors.New("distribution image not found")
}

func Build(dock *docker.Docker, opts *Options) error {
	log.Info("Building image")

	isImageBuilt, err := dock.IsImageBuilt(opts.Image.Name())
	if err != nil {
		return log.FailE(err)
	}
	if isImageBuilt && !opts.Rebuild {
		isImageOld, err := dock.IsImageOld(opts.Image.Name())
		if err != nil {
			return log.FailE(err)
		}
		if !isImageOld {
			return log.SkipE()
		}
	}

	repo, err := determineRepo(opts.Image.Tag())
	if err != nil {
		return err
	}

	log.Drop()

	args := docker.ImageBuildArgs{
		From: fmt.Sprintf("%s:%s", repo, opts.Image.Tag()),
		Name: opts.Image.Name(),
	}

	err = dock.ImageBuild(args)
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

// Create function commands Docker Engine to create and start container.
func Create(dock *docker.Docker, opts *Options) error {
	log.Info("Creating container")

	isContainerCreated, err := dock.IsContainerCreated(opts.Container.Name())
	if err != nil {
		return log.FailE(err)
	}
	if isContainerCreated {
		return log.SkipE()
	}

	args := docker.ContainerCreateArgs{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: opts.Dirs.Source.SourcePath(),
				Target: docker.ContainerSourceDir,
			}, {
				Type:   mount.TypeBind,
				Source: opts.Dirs.Build.ContainerPath(),
				Target: docker.ContainerBuildDir,
			}, {
				Type:   mount.TypeBind,
				Source: opts.Dirs.Cache.ImagePath(),
				Target: docker.ContainerCacheDir,
			},
		},
		Image: opts.Image.Name(),
		Name:  opts.Container.Name(),
		User:  fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
	}

	for _, mnt := range args.Mounts {
		err := os.MkdirAll(mnt.Source, os.ModePerm)
		if err != nil {
			return log.FailE(err)
		}
	}

	mounts, err := extraPackages(opts.ExtraPackages)
	if err != nil {
		return log.FailE(err)
	}
	args.Mounts = append(args.Mounts, mounts...)

	err = dock.ContainerCreate(args)
	if err != nil {
		return log.FailE(err)
	}

	err = dock.ContainerStart(opts.Container.Name())
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func extraPackages(packages []string) ([]mount.Mount, error) {
	mounts := make([]mount.Mount, 0)

	for _, pkg := range packages {
		source, err := filepath.Abs(pkg)
		if err != nil {
			return nil, err
		}

		info, err := os.Stat(source)
		if info == nil {
			return nil, err
		}
		if !info.IsDir() && !strings.HasSuffix(source, ".deb") {
			return nil, errors.New("please specify a directory or .deb file")
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

	return mounts, nil
}

func findTarball(pkg *naming.Package, in string) (string, error) {
	tarball := fmt.Sprintf(
		"%s/%s_%s.orig.tar",
		in,
		pkg.Source,
		pkg.Version.Version,
	)
	extensions := []string{".gz", ".xz", "bz2"}

	for _, ext := range extensions {
		info, _ := os.Stat(tarball + ext)
		if info != nil {
			return tarball + ext, nil
		}
	}

	return "", errors.New("tarball not found")
}

func moveTarball(tarball, dir string) error {
	source, err := filepath.EvalSymlinks(tarball)
	if err != nil {
		return err
	}

	target := filepath.Join(dir, filepath.Base(tarball))
	err = os.Rename(source, target)
	if err != nil {
		return err
	}

	return nil
}

func Package(dock *docker.Docker, opts *Options) error {
	log.Info("Packaging software")

	log.Drop()

	if !opts.Package.Version.IsNative() {
		tarball, err := findTarball(opts.Package, opts.Dirs.Build.ContainerPath())
		if err != nil {
			tarball, err = findTarball(opts.Package, opts.Dirs.Source.ParentPath())
			if err != nil {
				return log.FailE(err)
			}

			err = moveTarball(tarball, opts.Dirs.Build.ContainerPath())
			if err != nil {
				return log.FailE(err)
			}
		}
	}

	args := []docker.ContainerExecArgs{
		{
			Name:    opts.Container.Name(),
			Cmd:     "echo deb [trusted=yes] file://" + docker.ContainerArchiveDir + " ./ > a.list",
			WorkDir: "/etc/apt/sources.list.d",
		}, {
			Name:    opts.Container.Name(),
			Cmd:     "dpkg-scanpackages -m . > Packages",
			WorkDir: docker.ContainerArchiveDir,
		}, {
			Name:    opts.Container.Name(),
			Cmd:     "apt-get update",
			Network: true,
		}, {
			Name:    opts.Container.Name(),
			Cmd:     "apt-get build-dep",
			Network: true,
			WorkDir: docker.ContainerSourceDir,
		}, {
			Name:    opts.Container.Name(),
			Cmd:     "dpkg-buildpackage " + opts.DpkgFlags,
			Network: opts.Network,
		}, {
			Name: opts.Container.Name(),
			Cmd:  "debc",
		}, {
			Name:    opts.Container.Name(),
			Cmd:     "debi --with-depends",
			Network: true,
			AsRoot:  true,
		}, {
			Name: opts.Container.Name(),
			Cmd:  "lintian " + opts.LintianFlags,
		},
	}

	if opts.ExtraPackages == nil {
		args[0].Skip = true
		args[1].Skip = true
	}

	err := dock.ContainerNetwork(opts.Container.Name(), opts.Network)
	if err != nil {
		return log.FailE(err)
	}

	for _, arg := range args {
		err = dock.ContainerExec(arg)
		if err != nil {
			return log.FailE(err)
		}
	}

	return log.DoneE()
}

// Archive function moves successful build to archive by overwriting.
func Archive(dock *docker.Docker, opts *Options) error {
	log.Info("Archiving build")

	info, _ := os.Stat(opts.Dirs.Archive.PackageVersionPath())
	if info != nil {
		err := os.RemoveAll(opts.Dirs.Archive.PackageVersionPath())
		if err != nil {
			return log.FailE(err)
		}
	}

	err := os.MkdirAll(opts.Dirs.Archive.PackageSourcePath(), os.ModePerm)
	if err != nil {
		return log.FailE(err)
	}

	err = os.Rename(opts.Dirs.Build.ContainerPath(), opts.Dirs.Archive.PackageVersionPath())
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}

func removeAll(dock *docker.Docker) error {
	list, err := dock.ContainerList(app.Name)
	if err != nil {
		return log.FailE(err)
	}

	for _, container := range list {
		err = dock.ContainerRemove(container)
		if err != nil {
			return log.FailE(err)
		}
	}

	return nil
}

// Remove function commands Docker Engine to stop and remove container.
func Remove(dock *docker.Docker, opts *Options) error {
	log.Info("Removing container")

	if opts.All {
		return removeAll(dock)
	}

	isContainerCreated, err := dock.IsContainerCreated(opts.Container.Name())
	if err != nil {
		return log.FailE(err)
	}
	if !isContainerCreated {
		return log.SkipE()
	}

	isContainerStopped, err := dock.IsContainerStopped(opts.Container.Name())
	if err != nil {
		return log.FailE(err)
	}
	if !isContainerStopped {
		err = dock.ContainerStop(opts.Container.Name())
		if err != nil {
			return log.FailE(err)
		}
	}

	err = dock.ContainerRemove(opts.Container.Name())
	if err != nil {
		return log.FailE(err)
	}

	return log.DoneE()
}
