package util

import "github.com/docker/docker/api/types/mount"

func GetMounts() []mount.Mount {
	return nil
}

func CompareMounts(a, b []mount.Mount) bool {
	return true
}
