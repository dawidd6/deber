package stepping_test

import (
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
	}, {
		Name: "deps",
	}, {
		Name: "package",
	}, {
		Name: "test",
	}, {
		Name: "archive",
	}, {
		Name: "scan",
	}, {
		Name: "stop",
	}, {
		Name: "remove",
	},
}

var (
	includes = []string{"deps", "package", "scan"}
	excludes = []string{"deps", "package", "scan"}
)

func TestGet(t *testing.T) {
	included, excluded := steps.Get()

	if len(included) != len(steps) {
		t.Fatal("len mismatch")
	}
	if len(excluded) > 0 {
		t.Fatal("len > 0")
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
	len1 := len(included)
	len2 := len(includes)

	if len1 != len2 {
		t.Fatal("len mismatch", len1, len2)
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

	if len(excluded) != len(excludes) {
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

		if !found {
			t.Fatal("not found")
		}
	}
}

func TestReset(t *testing.T) {
	steps.Reset()
}

func TestGetAfterReset(t *testing.T) {
	TestGet(t)
}