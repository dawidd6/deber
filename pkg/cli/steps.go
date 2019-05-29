package cli

import (
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
	"github.com/docker/docker/api/types/mount"
	"os"
	"path/filepath"
	"strings"
)

func needBuild(dock *docker.Docker, name *naming.Naming) (bool, error) {
	isImageBuilt, err := dock.IsImageBuilt(name.Image.Name())
	if err != nil {
		return false, err
	}
	if isImageBuilt {
		isImageOld, err := dock.IsImageOld(name.Image.Name())
		if err != nil {
			return false, err
		}
		if isImageOld {
			return true, nil
		}

		return false, nil
	}

	return true, nil
}

func needCreate(dock *docker.Docker, name *naming.Naming) (bool, error) {
	isContainerCreated, err := dock.IsContainerCreated(name.Container.Name())
	if err != nil {
		return false, err
	}
	if isContainerCreated {
		return false, nil
	}

	return true, nil
}

func needRemove(dock *docker.Docker, name *naming.Naming) (bool, error) {
	isContainerCreated, err := dock.IsContainerCreated(name.Container.Name())
	if err != nil {
		return false, err
	}
	if isContainerCreated {
		return true, err
	}

	return false, nil
}

func runBuild(dock *docker.Docker, name *naming.Naming) error {
	repo, err := determineRepo(name.Image.Tag())
	if err != nil {
		return err
	}

	buildArgs := docker.ImageBuildArgs{
		From: fmt.Sprintf("%s:%s", repo, name.Image.Tag()),
		Name: name.Image.Name(),
	}

	err = dock.ImageBuild(buildArgs)
	if err != nil {
		return err
	}

	return nil
}

func runCreate(dock *docker.Docker, name *naming.Naming) error {
	createArgs := docker.ContainerCreateArgs{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: name.Dirs.Source.SourcePath(),
				Target: docker.ContainerSourceDir,
			}, {
				Type:   mount.TypeBind,
				Source: name.Dirs.Build.ContainerPath(),
				Target: docker.ContainerBuildDir,
			}, {
				Type:   mount.TypeBind,
				Source: name.Dirs.Cache.ImagePath(),
				Target: docker.ContainerCacheDir,
			},
		},
		Image: name.Image.Name(),
		Name:  name.Container.Name(),
		User:  fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
	}

	for _, mnt := range createArgs.Mounts {
		err := os.MkdirAll(mnt.Source, os.ModePerm)
		if err != nil {
			return err
		}
	}

	mounts, err := extraPackages()
	if err != nil {
		return err
	}
	createArgs.Mounts = append(createArgs.Mounts, mounts...)

	err = dock.ContainerCreate(createArgs)
	if err != nil {
		return err
	}

	return nil
}

func runTarball(name *naming.Naming) error {
	if !name.Package.Version.IsNative() {
		tarball, err := findTarball(name.Package, name.Dirs.Build.ContainerPath())
		if err != nil {
			tarball, err = findTarball(name.Package, name.Dirs.Source.ParentPath())
			if err != nil {
				return err
			}

			err = moveTarball(tarball, name.Dirs.Build.ContainerPath())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func runPackage(dock *docker.Docker, name *naming.Naming) error {
	execArgs := []docker.ContainerExecArgs{
		{
			Name:    name.Container.Name(),
			Cmd:     "echo deb [trusted=yes] file://" + docker.ContainerArchiveDir + " ./ > a.list",
			AsRoot:  true,
			WorkDir: "/etc/apt/sources.list.d",
		}, {
			Name:    name.Container.Name(),
			Cmd:     "dpkg-scanpackages -m . > Packages",
			AsRoot:  true,
			WorkDir: docker.ContainerArchiveDir,
		}, {
			Name:    name.Container.Name(),
			Cmd:     "apt-get update",
			AsRoot:  true,
			Network: true,
		}, {
			Name:    name.Container.Name(),
			Cmd:     "apt-get build-dep ./",
			Network: true,
			AsRoot:  true,
			WorkDir: docker.ContainerSourceDir,
		}, {
			Name:    name.Container.Name(),
			Cmd:     "dpkg-buildpackage " + flagDpkgFlags,
			Network: flagWithNetwork,
		}, {
			Name: name.Container.Name(),
			Cmd:  "debc",
		}, {
			Name:    name.Container.Name(),
			Cmd:     "debi --with-depends",
			Network: true,
			AsRoot:  true,
		}, {
			Name: name.Container.Name(),
			Cmd:  "lintian " + flagLintianFlags,
		},
	}

	if flagExtraPackages == nil {
		execArgs[0].Skip = true
		execArgs[1].Skip = true
	}

	err := dock.ContainerNetwork(name.Container.Name(), flagWithNetwork)
	if err != nil {
		return err
	}

	for _, arg := range execArgs {
		err = dock.ContainerExec(arg)
		if err != nil {
			return err
		}
	}

	return nil
}

func runRemove(dock *docker.Docker, name *naming.Naming) error {
	err := dock.ContainerRemove(name.Container.Name())
	if err != nil {
		return err
	}

	return err
}

func runArchive(name *naming.Naming) error {
	info, _ := os.Stat(name.Dirs.Archive.PackageVersionPath())
	if info != nil {
		err := os.RemoveAll(name.Dirs.Archive.PackageVersionPath())
		if err != nil {
			return err
		}
	}

	err := os.MkdirAll(name.Dirs.Archive.PackageSourcePath(), os.ModePerm)
	if err != nil {
		return err
	}

	err = os.Rename(name.Dirs.Build.ContainerPath(), name.Dirs.Archive.PackageVersionPath())
	if err != nil {
		return err
	}

	return nil
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

func extraPackages() ([]mount.Mount, error) {
	mounts := make([]mount.Mount, 0)

	for _, pkg := range flagExtraPackages {
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
