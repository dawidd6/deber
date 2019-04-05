package debian

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Debian struct {
	SourceName      string
	PackageVersion  string
	UpstreamVersion string
	TargetDist      string
	IsNative        bool
}

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

func New(line string) *Debian {
	return &Debian{
		SourceName:      SourceName(line),
		PackageVersion:  PackageVersion(line),
		UpstreamVersion: UpstreamVersion(line),
		TargetDist:      TargetDist(line),
		IsNative:        IsNative(line),
	}
}

func SourceName(line string) string {
	return strings.Split(line, " ")[0]
}

func PackageVersion(line string) string {
	packageVersion := strings.Split(line, " ")[1]
	packageVersion = strings.TrimPrefix(packageVersion, "(")
	packageVersion = strings.TrimSuffix(packageVersion, ")")

	return packageVersion
}

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

func TargetDist(line string) string {
	targetDist := strings.Split(line, " ")[2]
	targetDist = strings.TrimSuffix(targetDist, ";")

	if strings.Contains(targetDist, "-") {
		targetDist = strings.Split(targetDist, "-")[0]
	}

	return targetDist
}

func IsNative(line string) bool {
	version := strings.Split(line, " ")[1]

	if strings.Contains(version, "-") {
		return false
	}

	return true
}

func (debian *Debian) LocateTarball() (string, error) {
	if debian.IsNative {
		return "", nil
	}

	sourceName := debian.SourceName
	upstreamVersion := debian.UpstreamVersion
	tarball := fmt.Sprintf("%s_%s.orig.tar", sourceName, upstreamVersion)

	path, err := filepath.Abs(fmt.Sprintf("../%s", tarball))
	if err != nil {
		return "", err
	}

	for _, ext := range []string{".gz", ".xz", "bz2"} {
		info, _ := os.Stat(path + ext)
		if info != nil {
			return tarball + ext, nil
		}
	}

	return "", errors.New("tarball not found")
}
