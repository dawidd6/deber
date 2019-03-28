package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func getTarball(pkg, version string) (string, error) {
	compressions := []string{".gz", ".xz"}
	tarball := fmt.Sprintf("%s_%s.orig.tar", pkg, version)

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
