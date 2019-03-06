package app

type Step struct {
	label       string
	description string
	run         func() error
	disabled    bool
}

var steps = []Step{
	{
		label:       "build",
		description: "build Docker image",
		run:         runBuild,
	}, {
		label:       "create",
		description: "create Docker container",
		run:         runCreate,
	}, {
		label:       "start",
		description: "start Docker container to do nothing",
		run:         runStart,
	}, {
		label:       "package",
		description: "build Debian package in Docker container",
		run:         runPackage,
	}, {
		label:       "test",
		description: "test Debian package in Docker container",
		run:         runTest,
	}, {
		label:       "stop",
		description: "stop Docker container",
		run:         runStop,
	}, {
		label:       "remove",
		description: "remove Docker container",
		run:         runRemove,
	},
}
