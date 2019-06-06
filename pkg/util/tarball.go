package util

import (
	"fmt"
	"path/filepath"
	"pault.ag/go/debian/changelog"
	"strings"
)

func GetTarballBaseFileName(debian *changelog.ChangelogEntry) string {
	return fmt.Sprintf("%s_%s.orig.tar", debian.Source, debian.Version.Version)
}

func FindTarball(debian *changelog.ChangelogEntry, dir string) (string, bool) {
	tarball := GetTarballBaseFileName(debian)
	found := false

	err := Walk(dir, 1, func(node Node) bool {
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
