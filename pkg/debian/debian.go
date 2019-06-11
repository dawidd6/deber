package debian

import (
	"fmt"
	"github.com/dawidd6/deber/pkg/walk"
	"path/filepath"
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
	tarball := debian.GetTarballBaseFileName()
	found := false

	err := walk.Walk(dir, 1, func(node walk.Node) bool {
		file := filepath.Base(node.Path)
		if strings.HasPrefix(file, tarball) {
			tarball = file
			found = true
			return true
		}

		return false
	})
	if err != nil {
		return "", false
	}

	return tarball, found
}
