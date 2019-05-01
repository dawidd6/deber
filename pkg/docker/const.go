package docker

import "time"

const (
	ApiVersion = "1.30"

	MaxImageAge = time.Hour * 24 * 14

	ContainerStateRunning    = "running"
	ContainerStateCreated    = "created"
	ContainerStateExited     = "exited"
	ContainerStateRestarting = "restarting"
	ContainerStatePaused     = "paused"
	ContainerStateDead       = "dead"

	// Directories in container where their host counterpart should be mounted.
	ContainerArchiveDir = "/archive"
	ContainerBuildDir   = "/build"
	ContainerSourceDir  = "/build/source"
	ContainerCacheDir   = "/var/cache/apt"
)
