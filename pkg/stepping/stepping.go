package stepping

import "fmt"

type Step struct {
	Name        string
	Description []string
	Run         func() error
	excluded    bool
}

type Steps []*Step

var extraFunctionAfterRun func() error

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
			suggestion := steps.Suggest(name)
			return fmt.Errorf("step name \"%s\" is not valid, did you mean \"%s\"?", name, suggestion)
		}
	}

	return nil
}

func (steps Steps) countCharacters(s string) map[string]int {
	m := make(map[string]int)
	for _, char := range s {
		m[string(char)]++
	}

	return m
}

func (steps Steps) Suggest(name string) string {
	// returned match string
	match := ""

	maxMatch := 0
	maxHit := 0
	inputMap := steps.countCharacters(name)

	for _, step := range steps {
		currentMatch := 0
		currentHit := 0
		stepMap := steps.countCharacters(step.Name)

		// scan for matching characters
		for inputChar, inputCount := range inputMap {
			stepCount, ok := stepMap[inputChar]

			if ok {
				currentMatch++

				if stepCount == inputCount {
					currentHit++
				}
			}
		}

		// power up if last characters match
		if step.Name[len(step.Name)-1] == name[len(name)-1] {
			currentMatch++
		}

		// power up if first characters match
		if step.Name[0] == name[0] {
			currentMatch++
		}

		// check if there is a better match
		if maxHit < currentHit || maxMatch < currentMatch {
			maxHit = currentHit
			maxMatch = currentMatch
			match = step.Name
		}
	}

	return match
}

func (steps Steps) ExtraFunctionAfterRun(f func() error) {
	extraFunctionAfterRun = f
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

	steps.ExtraFunctionAfterRun(nil)
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

	if extraFunctionAfterRun != nil {
		return extraFunctionAfterRun()
	}

	return nil
}
