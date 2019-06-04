package config

type Config struct {
	DpkgFlags     string
	LintianFlags  string
	ExtraPackages []string

	ArchiveBaseDir string
	CacheBaseDir   string
	BuildBaseDir   string
	SourceBaseDir  string
}
