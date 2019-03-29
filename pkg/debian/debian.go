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
	TarballFileName string
	IsNative        bool
}

func ParseFile() (*Debian, error) {
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

	debian := ParseLine(line)

	if !debian.IsNative {
		tarball, err := debian.FindTarball()
		if err != nil {
			return nil, err
		}

		debian.TarballFileName = tarball
	}

	return debian, nil
}

func ParseLine(line string) *Debian {
	elements := strings.Split(line, " ")

	sourceName := elements[0]

	packageVersion := elements[1]
	packageVersion = strings.TrimPrefix(packageVersion, "(")
	packageVersion = strings.TrimSuffix(packageVersion, ")")
	isNative := true

	upstreamVersion := packageVersion
	if strings.Contains(upstreamVersion, ":") {
		upstreamVersion = strings.Split(upstreamVersion, ":")[1]
	}
	if strings.Contains(upstreamVersion, "-") {
		upstreamVersion = strings.Split(upstreamVersion, "-")[0]
		isNative = false
	}

	targetDist := elements[2]
	targetDist = strings.TrimSuffix(targetDist, ";")
	if strings.Contains(targetDist, "-") {
		targetDist = strings.Split(targetDist, "-")[0]
	}

	return &Debian{
		SourceName:      sourceName,
		PackageVersion:  packageVersion,
		UpstreamVersion: upstreamVersion,
		TargetDist:      targetDist,
		IsNative:        isNative,
	}
}

func (debian *Debian) FindTarball() (string, error) {
	tarball := fmt.Sprintf("%s_%s.orig.tar", debian.SourceName, debian.UpstreamVersion)

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
