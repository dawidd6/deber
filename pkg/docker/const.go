package docker

import "time"

const (
	// ApiVersion constant is the minimum supported version of Docker Engine API
	ApiVersion = "1.30"

	// ImageMaxAge constant defines how old image can be
	// If image was created ImageMaxAge time ago, then
	// it should be rebuilt
	ImageMaxAge = time.Hour * 24 * 14

	// ContainerStopTimeout constant represents how long Docker Engine
	// will wait for container before stopping it
	ContainerStopTimeout = time.Millisecond * 10

	// ContainerState* constants define various states of Docker container life
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
