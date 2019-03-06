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
		label: "create",
		run:   runCreate,
	}, {
		label: "start",
		run:   runStart,
	}, {
		label: "package",
		run:   runPackage,
	}, {
		label: "test",
		run:   runTest,
	}, {
		label: "stop",
		run:   runStop,
	}, {
		label: "remove",
		run:   runRemove,
	},
}
