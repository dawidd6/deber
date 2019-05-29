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
		newMounts, err := getMounts(name)
		if err != nil {
			return false, err
		}

		oldMounts, err := dock.ContainerMounts(name.Container.Name())
		if err != nil {
			return false, err
		}

		if len(newMounts) != len(oldMounts) {
			return true, nil
		}

		for _, newMount := range newMounts {
			found := false

			for _, oldMount := range oldMounts {
				if oldMount == newMount {
					found = true
				}
			}

			if !found {
				return true, nil
			}
		}
	} else {
		return true, nil
	}

	return false, nil
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
	repos := []string{"debian", "ubuntu"}
	repo, err := docker.MatchRepo(repos, name.Image.Tag())
	if err != nil {
		return err
	}

	err = dock.ImageBuild(
		docker.ImageBuildArgs{
			From: fmt.Sprintf("%s:%s", repo, name.Image.Tag()),
			Name: name.Image.Name(),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func getMounts(name *naming.Naming) ([]mount.Mount, error) {
	mounts := []mount.Mount{
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
	}

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

func runCreate(dock *docker.Docker, name *naming.Naming) error {
	mounts, err := getMounts(name)
	if err != nil {
		return err
	}

	for _, mnt := range mounts {
		info, _ := os.Stat(mnt.Source)
		if info != nil && !info.IsDir() {
			continue
		}

		err := os.MkdirAll(mnt.Source, os.ModePerm)
		if err != nil {
			return err
		}
	}

	err = dock.ContainerCreate(
		docker.ContainerCreateArgs{
			Mounts: mounts,
			Image:  name.Image.Name(),
			Name:   name.Container.Name(),
			User:   fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func runTarball(dock *docker.Docker, name *naming.Naming) error {
	if name.Package.Version.IsNative() {
		return nil
	}

	tarball := fmt.Sprintf(
		"%s_%s.orig.tar",
		name.Package.Source,
		name.Package.Version.Version,
	)
	extensions := []string{".gz", ".xz", "bz2"}
	dirs := []string{name.Dirs.Build.ContainerPath(), name.Dirs.Source.ParentPath()}

	for _, dir := range dirs {
		for _, ext := range extensions {
			source := filepath.Join(dir, tarball+ext)
			target := filepath.Join(dir, tarball+ext)

			info, _ := os.Stat(source)
			if info == nil {
				continue
			}

			source, err := filepath.EvalSymlinks(source)
			if err != nil {
				return err
			}

			err = os.Rename(source, target)
			if err != nil {
				return err
			}

			return nil
		}
	}

	return errors.New("tarball not found")
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
			Name: name.Container.Name(),
			Cmd:  "dpkg-buildpackage " + flagDpkgFlags,
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

	for _, arg := range execArgs {
		err := dock.ContainerExec(arg)
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

func runArchive(dock *docker.Docker, name *naming.Naming) error {
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
