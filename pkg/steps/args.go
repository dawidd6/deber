package steps

import "github.com/dawidd6/deber/pkg/naming"

type BuildArgs struct {
	*naming.Naming
	IsRebuildNeeded bool
}

type CreateArgs struct {
	*naming.Naming
	ExtraPackages []string
}

type NetworkArgs struct {
	*naming.Naming
	IsConnectionNeeded bool
}

type DependsArgs struct {
	*naming.Naming
	ExtraPackages []string
}

type PackageArgs struct {
	*naming.Naming
	DpkgFlags        string
	LintianFlags     string
	IsTestNeeded     bool
	IsNetworkNeeded  bool
	TarballSourceDir string
	TarballTargetDir string
}

type ArchiveArgs struct {
	*naming.Naming
}

type RemoveArgs struct {
	*naming.Naming
	IsAllSelected bool
}

type ShellArgs struct {
	*naming.Naming
}
