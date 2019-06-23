package naming

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
	"os"
	"path/filepath"
	"strings"
)

var (
	ArchiveBaseDir = filepath.Join(os.Getenv("HOME"), app.Name)
	BuildBaseDir   = "/tmp"
	CacheBaseDir   = "/tmp"
	SourceBaseDir  = os.Getenv("PWD")

	PackageName     = ""
	PackageVersion  = ""
	PackageUpstream = ""
	PackageTarget   = ""
)

func Image() string {
	return fmt.Sprintf(
		"%s:%s",
		app.Name,
		standardizeImageTag(),
	)
}

func Container() string {
	return fmt.Sprintf(
		"%s_%s_%s_%s",
		app.Name,
		PackageTarget,
		PackageName,
		standardizePackageVersion(PackageVersion),
	)
}

func standardizePackageVersion(version string) string {
	// Docker allows only [a-zA-Z0-9][a-zA-Z0-9_.-]
	// and Debian package versioning allows these characters
	version = strings.Replace(version, "~", "-", -1)
	version = strings.Replace(version, ":", "-", -1)
	version = strings.Replace(version, "+", "-", -1)

	return version
}

func standardizeImageTag() string {
	if strings.Contains(PackageVersion, "bpo") {
		if strings.Contains(PackageTarget, "backports") {
			return PackageTarget
		}

		if PackageTarget == "UNRELEASED" {
			return "unstable"
		}
	}

	if strings.Contains(PackageTarget, "-") {
		return strings.Split(PackageTarget, "-")[0]
	}

	return PackageTarget
}

func BuildDir() string {
	return filepath.Join(
		BuildBaseDir,
		Container(),
	)
}

func CacheDir() string {
	return filepath.Join(
		CacheBaseDir,
		Image(),
	)
}

func ArchiveTargetDir() string {
	return filepath.Join(
		ArchiveBaseDir,
		PackageTarget,
	)
}

func ArchivePackageDir() string {
	return filepath.Join(
		ArchiveTargetDir(),
		PackageName,
	)
}

func ArchiveVersionDir() string {
	return filepath.Join(
		ArchivePackageDir(),
		PackageVersion,
	)
}

func SourceDir() string {
	return SourceBaseDir
}

func SourceParentDir() string {
	return filepath.Dir(
		SourceDir(),
	)
}
