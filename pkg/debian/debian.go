package debian

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"pault.ag/go/debian/changelog"
)

type Debian struct {
	Source   string
	Version  string
	Upstream string
	Native   bool
	Tarball  string
}

func New() (*Debian, error) {
	ch, err := changelog.ParseFileOne("debian/changelog")
	if err != nil {
		return nil, err
	}

	tarball, err := getTarball(ch)
	if !ch.Version.IsNative() {
		if err != nil {
			return nil, err
		}
	}

	return &Debian{
		Source:   ch.Source,
		Version:  ch.Version.String(),
		Upstream: ch.Version.Version,
		Native:   ch.Version.IsNative(),
		Tarball:  tarball,
	}, nil
}

func getTarball(ch *changelog.ChangelogEntry) (string, error) {
	if ch.Version.IsNative() {
		return "", nil
	}

	compressions := []string{".gz", ".xz"}
	tarball := fmt.Sprintf("%s_%s.orig.tar", ch.Source, ch.Version.Version)

	path, err := filepath.Abs(fmt.Sprintf("../%s", tarball))
	if err != nil {
		return "", err
	}

	for i := range compressions {
		stat, _ := os.Stat(path + compressions[i])
		if stat != nil {
			tarball += compressions[i]
			return tarball, nil
		}
	}

	return "", errors.New("tarball not found")
}
