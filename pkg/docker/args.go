package docker

type BuildImageArgs struct {
	// Full parent image name,
	// placed in Dockerfile's FROM directive
	//
	// Example: ubuntu:bionic
	From string
	// Full to-be-built image name
	//
	// Example: deber:unstable
	Name string
}

type ContainerCreateArgs struct {
	SourceDir  string
	BuildDir   string
	ArchiveDir string
	CacheDir   string
	Image      string
	Name       string
}

type ContainerExecArgs struct {
	Interactive bool
	Name        string
	Cmd         string
	WorkDir     string
}

type ContainerExecResizeArgs struct {
	Fd     uintptr
	ExecID string
}

type DockerfileArgs struct {
}
