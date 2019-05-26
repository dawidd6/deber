// Package debian holds informations about Debian package.
package debian

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Debian struct represents some informations about package.
type Debian struct {
	SourceName      string
	PackageVersion  string
	UpstreamVersion string
	TargetDist      string
	IsNative        bool
}

// ParseChangelog reads the contents of debian/changelog file
// and parses it to Debian struct.
func ParseChangelog() (*Debian, error) {
	file, err := os.Open("debian/changelog")
	if err != nil {
		return nil, err
	}

	line, err := bufio.NewReader(file).ReadString('\n')
	if err != nil {
		return nil, err
	}

	err = file.Close()
	if err != nil {
		return nil, err
	}

	return New(line), nil
}

// New creates a fresh Debian struct with fields parsed
// from single changelog line.
func New(line string) *Debian {
	return &Debian{
		SourceName:      SourceName(line),
		PackageVersion:  PackageVersion(line),
		UpstreamVersion: UpstreamVersion(line),
		TargetDist:      TargetDist(line),
		IsNative:        IsNative(line),
	}
}

// SourceName parses single changelog line to extract
// source package name.
func SourceName(line string) string {
	return strings.Split(line, " ")[0]
}

// PackageVersion parses single changelog line to extract
// source package version.
func PackageVersion(line string) string {
	packageVersion := strings.Split(line, " ")[1]
	packageVersion = strings.TrimPrefix(packageVersion, "(")
	packageVersion = strings.TrimSuffix(packageVersion, ")")

	return packageVersion
}

// UpstreamVersion parses single changelog line to extract
// upstream source version.
func UpstreamVersion(line string) string {
	upstreamVersion := PackageVersion(line)

	if strings.Contains(upstreamVersion, ":") {
		upstreamVersion = strings.Split(upstreamVersion, ":")[1]
	}

	if strings.Contains(upstreamVersion, "-") {
		upstreamVersion = strings.Split(upstreamVersion, "-")[0]
	}

	return upstreamVersion
}

// TargetDist parses single changelog line to extract
// target distribution.
func TargetDist(line string) string {
	targetDist := strings.Split(line, " ")[2]
	targetDist = strings.TrimSuffix(targetDist, ";")

	if strings.Contains(targetDist, "-") {
		targetDist = strings.Split(targetDist, "-")[0]
	}

	// Debian backport
	if strings.Contains(PackageVersion(line), "bpo") {
		targetDist += "-backports"
	}

	if targetDist == "UNRELEASED" {
		targetDist = "unstable"
	}

	return targetDist
}

// IsNative checks if package is native by searching for single '-'
// in package's version string.
func IsNative(line string) bool {
	version := strings.Split(line, " ")[1]

	if strings.Contains(version, "-") {
		return false
	}

	return true
}

// LocateTarball searches parent directory for orig upstream tarball
// and returns the complete filename of it, not filepath.
func (debian *Debian) LocateTarball(dir string) (string, error) {
	if debian.IsNative {
		return "", nil
	}

	sourceName := debian.SourceName
	upstreamVersion := debian.UpstreamVersion
	tarball := fmt.Sprintf("%s_%s.orig.tar", sourceName, upstreamVersion)

	path := filepath.Join(dir, tarball)

	for _, ext := range []string{".gz", ".xz", "bz2"} {
		info, _ := os.Stat(path + ext)
		if info != nil {
			return tarball + ext, nil
		}
	}

	return "", errors.New("tarball not found")
}
