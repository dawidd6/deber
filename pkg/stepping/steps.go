package stepping

type Steps []*Step

func (steps Steps) IsNameValid(name string) bool {
	found := false

	steps.Walk(func(step *Step) {
		if name == step.Name {
			found = true
			return
		}
	})

	return found
}

func (steps Steps) Run() error {
	for _, step := range steps {
		err := step.Func()
		if err != nil {
			return err
		}
	}

	return nil
}

func (steps Steps) Include(names ...string) Steps {
	onlySteps := make(Steps, 0)

	steps.Walk(func(step *Step) {
		for _, name := range names {
			if name == step.Name {
				onlySteps = append(onlySteps, step)
				return
			}
		}
	})

	return onlySteps
}

func (steps Steps) Exclude(names ...string) Steps {
	onlySteps := make(Steps, 0)
	found := false

	steps.Walk(func(step *Step) {
		found = false

		for _, name := range names {
			if name == step.Name {
				found = true
				break
			}
		}

		if !found {
			onlySteps = append(onlySteps, step)
		}
	})

	return onlySteps
}

func (steps Steps) Required() Steps {
	onlySteps := make(Steps, 0)

	steps.Walk(func(step *Step) {
		if !step.Optional {
			onlySteps = append(onlySteps, step)
		}
	})

	return onlySteps
}

func (steps Steps) Walk(f func(step *Step)) {
	for _, step := range steps {
		f(step)
	}
}
