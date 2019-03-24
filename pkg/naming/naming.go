package naming

import (
	"fmt"
	"os"
	"strings"
)

const (
	ContainerRepoDir   = "/repo"
	ContainerBuildDir  = "/build"
	ContainerSourceDir = "/build/source"
	ContainerCacheDir  = "/var/cache/apt"
)

func Container(program, image, source, version string) string {
	// Docker allows only [a-zA-Z0-9][a-zA-Z0-9_.-]
	// and Debian versioning allows below characters
	version = strings.Replace(version, "~", "-", -1)
	version = strings.Replace(version, ":", "-", -1)
	version = strings.Replace(version, "+", "-", -1)
	image = strings.Replace(image, ":", "-", -1)
	image = strings.Replace(image, "/", "-", -1)

	return fmt.Sprintf(
		"%s_%s_%s-%s",
		program,
		image,
		source,
		version,
	)
}

func Image(program, image string) string {
	return fmt.Sprintf(
		"%s-%s",
		program,
		image,
	)
}

func HostCacheDir(image string) string {
	return fmt.Sprintf(
		"/tmp/%s",
		image,
	)
}

func HostSourceDir() string {
	return os.Getenv("PWD")
}

func HostBuildDir(container string) string {
	return fmt.Sprintf(
		"%s/../%s",
		HostSourceDir(),
		container,
	)
}

func SourceTarball(tarball string) string {
	return fmt.Sprintf(
		"%s/../%s",
		HostSourceDir(),
		tarball,
	)
}

func TargetTarball(container, tarball string) string {
	return fmt.Sprintf(
		"%s/%s",
		HostBuildDir(container),
		tarball,
	)
}
