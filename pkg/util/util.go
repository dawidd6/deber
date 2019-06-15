package util

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func CopyFile(source, target string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}

	targetFile, err := os.Create(target)
	if err != nil {
		return err
	}

	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

func CopyDir(source, target string, overwrite bool) error {
	stat, _ := os.Stat(target)
	if stat != nil && overwrite {
		err := os.Remove(target)
		if err != nil {
			return err
		}
	}

	err := os.Mkdir(target, os.ModePerm)
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(source)
	if err != nil {
		return err
	}

	for _, file := range files {
		src := filepath.Join(source, file.Name())
		tar := filepath.Join(target, file.Name())

		if file.IsDir() {
			err = CopyDir(file.Name(), filepath.Join(), overwrite)
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(src, tar)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
