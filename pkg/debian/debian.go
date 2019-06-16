package debian

import (
	"fmt"
	"io/ioutil"
	"pault.ag/go/debian/changelog"
	"strings"
)

var ChangelogPath = "debian/changelog"

type Debian struct {
	Source  string
	Version Version
	Target  string
}

type Version struct {
	Package  string
	Upstream string
	Native   bool
}

func New() (*Debian, error) {
	ch, err := changelog.ParseFileOne(ChangelogPath)
	if err != nil {
		return nil, err
	}

	return &Debian{
		Source: ch.Source,
		Target: ch.Target,
		Version: Version{
			Package:  ch.Version.String(),
			Upstream: ch.Version.Version,
			Native:   ch.Version.IsNative(),
		},
	}, nil
}

func (debian *Debian) GetTarballBaseFileName() string {
	return fmt.Sprintf("%s_%s.orig.tar", debian.Source, debian.Version.Upstream)
}

func (debian *Debian) FindTarball(dir string) (string, bool) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", false
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), debian.GetTarballBaseFileName()) {
			return file.Name(), true
		}
	}

	return "", false
}
