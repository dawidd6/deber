package main

type Step struct {
	label    string
	run      func() error
	disabled bool
}

var steps = []Step{
	{
		label: "build",
		run:   runBuild,
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
