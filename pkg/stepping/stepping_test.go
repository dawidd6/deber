package stepping_test

import (
	"errors"
	"github.com/dawidd6/deber/pkg/debian"
	"github.com/dawidd6/deber/pkg/docker"
	"github.com/dawidd6/deber/pkg/naming"
	"testing"

	"github.com/dawidd6/deber/pkg/stepping"
)

var steps = stepping.Steps{
	{
		Name: "check",
	}, {
		Name: "build",
	}, {
		Name: "create",
	}, {
		Name: "start",
	}, {
		Name: "tarball",
	}, {
		Name: "update",
		Run:  runNil,
	}, {
		Name: "deps",
		Run:  runError,
	}, {
		Name: "package",
		Run:  runNil,
	}, {
		Name: "test",
	}, {
		Name: "archive",
		Run:  runNil,
	}, {
		Name: "scan",
		Run:  runError,
	}, {
		Name: "stop",
	}, {
		Name: "remove",
	}, {
		Name:     "shell",
		Optional: true,
		Excluded: true,
	},
}

var (
	includes = []string{"archive", "package", "update"}
	excludes = []string{"archive", "package", "update"}
)

func runNil(*debian.Debian, *docker.Docker, *naming.Naming) error {
	return nil
}

func runError(*debian.Debian, *docker.Docker, *naming.Naming) error {
	return errors.New("error")
}

func TestSuggest(t *testing.T) {
	cases := map[string]string{
		"eps":         "deps",
		"udate":       "update",
		"date":        "update",
		"move":        "remove",
		"stap":        "stop",
		"stahp":       "stop",
		"buld":        "build",
		"arhiche":     "archive",
		"archibald":   "archive",
		"buildah":     "build",
		"depends":     "deps",
		"uil":         "build",
		"ate":         "update",
		"up":          "update",
		"rm":          "remove",
		"dep":         "deps",
		"scam":        "scan",
		"tar":         "tarball",
		"checkyoself": "check",
		"pkg":         "package",
		"sto":         "stop",
		"sta":         "start",
	}

	for gave, expected := range cases {
		got := steps.Suggest(gave)
		if got != expected {
			t.Fatal("gave:", gave, "expected:", expected, "got:", got)
		}
	}
}

func TestGet(t *testing.T) {
	included, excluded := steps.Get()

	// minus one optional step
	if len(included) != len(steps)-1 {
		t.Fatal("len mismatch")
	}

	// one optional step
	if len(excluded) != 1 {
		t.Fatal("len mismatch")
	}
}

func TestInclude(t *testing.T) {
	err := steps.Include(includes...)
	if err != nil {
		t.Fatal(err)
	}

	err = steps.Include()
	if err != nil {
		t.Fatal(err)
	}

	err = steps.Include("some", "body", "once", "told", "me")
	if err == nil {
		t.Fatal(err)
	}

	err = steps.Include("some,comma")
	if err == nil {
		t.Fatal(err)
	}
}

func TestGetAfterInclude(t *testing.T) {
	included, _ := steps.Get()

	if len(included) != len(includes) {
		t.Fatal("len mismatch")
	}

	for _, i := range included {
		found := false

		for _, j := range includes {
			if j == i.Name {
				found = true
				break
			}
		}

		if !found {
			t.Fatal("not found")
		}
	}
}

func TestRunAfterInclude(t *testing.T) {
	err := steps.Run(nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExclude(t *testing.T) {
	err := steps.Exclude(excludes...)
	if err != nil {
		t.Fatal(err)
	}

	err = steps.Exclude()
	if err != nil {
		t.Fatal(err)
	}

	err = steps.Exclude("some", "body", "once", "told", "me")
	if err == nil {
		t.Fatal(err)
	}

	err = steps.Exclude("some,comma")
	if err == nil {
		t.Fatal(err)
	}
}

func TestGetAfterExclude(t *testing.T) {
	_, excluded := steps.Get()

	// plus one optional step
	if len(excluded) != len(excludes)+1 {
		t.Fatal("len mismatch")
	}

	for _, i := range excluded {
		found := false

		for _, j := range excludes {
			if j == i.Name {
				found = true
				break
			}
		}

		if !found && !i.Optional {
			t.Fatal("not found")
		}
	}
}

func TestRunAfterExclude(t *testing.T) {
	err := steps.Run(nil, nil, nil)
	if err == nil {
		t.Fatal(err)
	}
}

func TestReset(t *testing.T) {
	steps.Reset()
}

func TestGetAfterReset(t *testing.T) {
	TestGet(t)
}
