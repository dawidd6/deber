package stepping

import "fmt"

type Step struct {
	Name        string
	Description string
	Run         func() error
	excluded    bool
}

type Steps []*Step

func (steps Steps) isNameValid(name string) bool {
	for _, step := range steps {
		if step.Name == name {
			return true
		}
	}

	return false
}

func (steps Steps) validateNames(names ...string) error {
	for _, name := range names {
		if !steps.isNameValid(name) {
			return fmt.Errorf("step name \"%s\" is not valid", name)
		}
	}

	return nil
}

func (steps Steps) Include(names ...string) error {
	if len(names) == 0 {
		return nil
	}

	err := steps.validateNames(names...)
	if err != nil {
		return err
	}

	for _, step := range steps {
		step.excluded = true

		for _, name := range names {
			if name == step.Name {
				step.excluded = false
				break
			}
		}
	}

	return nil
}

func (steps Steps) Exclude(names ...string) error {
	if len(names) == 0 {
		return nil
	}

	err := steps.validateNames(names...)
	if err != nil {
		return err
	}

	for _, step := range steps {
		step.excluded = false

		for _, name := range names {
			if name == step.Name {
				step.excluded = true
				break
			}
		}
	}

	return nil
}

func (steps Steps) Get() (included, excluded Steps) {
	included = Steps{}
	excluded = Steps{}

	for _, step := range steps {
		if step.excluded {
			excluded = append(excluded, step)
		} else {
			included = append(included, step)
		}
	}

	return included, excluded
}

func (steps Steps) Reset() {
	for _, step := range steps {
		step.excluded = false
	}
}

func (steps Steps) Run() error {
	for _, step := range steps {
		if step.excluded {
			continue
		}

		err := step.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
