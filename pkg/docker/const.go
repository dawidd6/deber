package docker

import "time"

const (
	// APIVersion constant is the minimum supported version of Docker Engine API
	APIVersion = "1.30"

	// ImageMaxAge constant defines how old image can be
	//
	// If image was created ImageMaxAge time ago, then
	// it should be rebuilt
	ImageMaxAge = time.Hour * 24 * 14

	// ContainerStateRunning constants defines that container is running
	ContainerStateRunning = "running"
	// ContainerStateCreated constants defines that container is created
	ContainerStateCreated = "created"
	// ContainerStateExited constants defines that container has exited
	ContainerStateExited = "exited"
	// ContainerStateRestarting constants defines that container is restarting
	ContainerStateRestarting = "restarting"
	// ContainerStatePaused constants defines that container is paused
	ContainerStatePaused = "paused"
	// ContainerStateDead constants defines that container is dead
	ContainerStateDead = "dead"

	// ContainerArchiveDir constant represents where on container will
	// archive directory be mounted
	ContainerArchiveDir = "/archive"
	// ContainerBuildDir constant represents where on container will
	// build directory be mounted
	ContainerBuildDir = "/build"
	// ContainerSourceDir constant represents where on container will
	// source directory be mounted
	ContainerSourceDir = "/build/source"
	// ContainerCacheDir constant represents where on container will
	// cache directory be mounted
	ContainerCacheDir = "/var/cache/apt"
)
