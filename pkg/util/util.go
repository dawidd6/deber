// Package util includes useful helper functions
package util

import "github.com/docker/docker/api/types/mount"

// CompareMounts function simply compares if given mounts are equal
func CompareMounts(a, b []mount.Mount) bool {
	if len(a) != len(b) {
		return false
	}

	matches := 0
	for _, aMount := range a {
		for _, bMount := range b {
			if aMount == bMount {
				matches++
				break
			}
		}
	}

	if matches == len(a) {
		return true
	}

	return false
}
