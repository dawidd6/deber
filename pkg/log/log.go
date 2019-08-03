package log

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	cyan   = "\033[0;36m"
	blue   = "\033[0;34m"
	red    = "\033[0;31m"
	normal = "\033[0m"
)

var (
	NoColor bool
	Prefix  string
	dropped bool
)

func init() {
	Prefix = filepath.Base(os.Args[0])
}

func Drop() {
	if dropped {
		return
	}

	dropped = true
	fmt.Println()
}

func Info(info string) {
	dropped = false

	if NoColor {
		fmt.Printf("%s:info: %s ...", Prefix, info)
	} else {
		fmt.Printf("%s%s:info:%s %s ...", blue, Prefix, normal, info)
	}
}

func Error(err error) {
	if NoColor {
		fmt.Printf("%s:error: %s\n", Prefix, err)
	} else {
		fmt.Printf("%s%s:error:%s %s\n", red, Prefix, normal, err)
	}
}

func ExtraInfo(info string) {
	dropped = false
	fmt.Printf("  %s ...", info)
}

func Skipped() error {
	if !dropped {
		fmt.Printf("%s", "skipped")
		Drop()
	}

	return nil
}

func Done() error {
	if !dropped {
		fmt.Printf("%s", "done")
		Drop()
	}

	return nil
}

func Failed(err error) error {
	if !dropped {
		fmt.Printf("%s", "failed")
		Drop()
	}

	return err
}
