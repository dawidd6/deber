package steps

type BuildArgs struct {
	ImageName    string
	Distribution string
	Rebuild      bool
}

type CreateArgs struct {
	ImageName     string
	ContainerName string
	SourceDir     string
	BuildDir      string
	ArchiveDir    string
	CacheDir      string
	ExtraPackages []string
}

type TarballArgs struct {
	PackageName            string
	PackageUpstreamVersion string
	IsPackageNative        bool
	SourceDir              string
	TargetDir              string
}

type NetworkArgs struct {
	ContainerName string
	IsConnected   bool
}

type DependsArgs struct {
	ContainerName   string
	IsArchiveNeeded bool
}

type PackageArgs struct {
	ContainerName   string
	DpkgFlags       string
	LintianFlags    string
	IsTestNeeded    bool
	IsNetworkNeeded bool
}

type ArchiveArgs struct {
	ArchivePackageDir string
	BuildDir          string
}

type RemoveArgs struct {
	ContainerName string
}

type ShellArgs struct {
	ContainerName string
}
