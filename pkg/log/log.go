package log

import (
	"errors"
	"fmt"
	"github.com/dawidd6/deber/pkg/app"
)

const (
	blue   = "\033[0;34m"
	red    = "\033[0;31m"
	normal = "\033[0m"
)

var (
	Skip    = errors.New("SKIP")
	NoColor = false
	newLine = true
)

func Drop() {
	newLine = true
	fmt.Println()
}

func Result(err error) error {
	if newLine {
		return err
	}

	switch err {
	case Skip:
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

func Info(v interface{}) {
	if !NoColor {
		fmt.Printf("%s%s:info:%s %s ...", blue, app.Name, normal, v)
	} else {
		fmt.Printf("%s:info: %s ...", app.Name, v)
	}

	newLine = false
}

func Error(v interface{}) {
	if !NoColor {
		fmt.Printf("%s%s:error:%s %s\n", red, app.Name, normal, v)
	} else {
		fmt.Printf("%s:error: %s\n", app.Name, v)
	}
}
