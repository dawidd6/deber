package util

import (
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
	"os"
	"path/filepath"
)

func FindTarball(a *app.App) (string, string, error) {
	if a.Version.IsNative() {
		return "", "", nil
	}

	file := fmt.Sprintf("%s_%s.orig.tar", a.Source, a.Version.Version)
	extensions := []string{".gz", ".xz", "bz2"}
	dirs := []string{a.BuildDir(), a.SourceParentDir()}

	for _, dir := range dirs {
		for _, ext := range extensions {
			path := filepath.Join(dir, file+ext)

			info, _ := os.Stat(path)
			if info != nil {
				return file + ext, dir, nil
			}
		}
	}

	return "", "", errors.New("tarball not found")
}
