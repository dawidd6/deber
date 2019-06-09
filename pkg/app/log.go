package app

import (
	"errors"
	"fmt"
)

const (
	blue   = "\033[0;34m"
	red    = "\033[0;31m"
	normal = "\033[0m"
)

var (
	logSkip = errors.New("SKIP")
	newLine = true
)

func (a *App) LogDrop() {
	newLine = true
	fmt.Println()
}

func (a *App) LogResult(err error) error {
	if newLine {
		return err
	}

	switch err {
	case logSkip:
		fmt.Printf("skipped\n")
		return nil
	case nil:
		fmt.Printf("done\n")
		return nil
	default:
		fmt.Printf("failed\n")
		return err
	}
}

func (a *App) LogInfo(v interface{}) {
	if !a.Config.LogNoColor {
		fmt.Printf("%s%s:info:%s %s ...", blue, a.Name, normal, v)
	} else {
		fmt.Printf("%s:info: %s ...", a.Name, v)
	}

	newLine = false
}

func (a *App) LogError(v interface{}) {
	if !a.Config.LogNoColor {
		fmt.Printf("%s%s:error:%s %s\n", red, a.Name, normal, v)
	} else {
		fmt.Printf("%s:error: %s\n", a.Name, v)
	}
}
