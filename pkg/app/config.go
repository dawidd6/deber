package app

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	DpkgFlags     string `yaml:"dpkg-flags"`
	LintianFlags  string
	Changelog     string
	ExtraPackages []string `yaml:"-"`
	LogNoColor    bool
	NoRebuild     bool
	Start         bool   `yaml:"-"`
	Stop          bool   `yaml:"-"`
	Dist          string `yaml:"-"`

	ArchiveBaseDir string
	CacheBaseDir   string
	BuildBaseDir   string
	SourceBaseDir  string
}

func (a *App) DefaultConfig() *Config {
	return &Config{
		DpkgFlags:    "-tc",
		LintianFlags: "-i -I",
		Changelog:    "debian/changelog",

		ArchiveBaseDir: filepath.Join(os.Getenv("HOME"), a.Name),
		CacheBaseDir:   "/tmp",
		BuildBaseDir:   "/tmp",
		SourceBaseDir:  os.Getenv("PWD"),
	}
}

func (a *App) ConfigFile() string {
	return filepath.Join(
		os.Getenv("HOME"),
		".config",
		a.Name,
		"config.yaml",
	)
}

func (a *App) Configure() {
	a.Config = a.DefaultConfig()

	bytes, _ := ioutil.ReadFile(a.ConfigFile())
	_ = yaml.Unmarshal(bytes, a.Config)
}
