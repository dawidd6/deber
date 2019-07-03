package logger

import (
	"fmt"
)

const (
	cyan   = "\033[0;36m"
	blue   = "\033[0;34m"
	red    = "\033[0;31m"
	normal = "\033[0m"
)

type Logger struct {
	prefix  string
	color   bool
	dropped bool
}

func New(prefix string, color bool) *Logger {
	return &Logger{
		prefix: prefix,
		color:  color,
	}
}

func (log *Logger) Drop() {
	if log.dropped {
		return
	}

	log.dropped = true
	fmt.Println()
}

func (log *Logger) Info(info string) {
	log.dropped = false

	if log.color {
		fmt.Printf("%s%s:info:%s %s ...", blue, log.prefix, normal, info)
	} else {
		fmt.Printf("%s:info: %s ...", log.prefix, info)
	}
}

func (log *Logger) Error(err error) {
	if log.color {
		fmt.Printf("%s%s:error:%s %s\n", red, log.prefix, normal, err)
	} else {
		fmt.Printf("%s:error: %s\n", log.prefix, err)
	}
}

func (log *Logger) ExtraInfo(info string) {
	log.dropped = false
	fmt.Printf("  %s ...", info)
}

func (log *Logger) Skipped() error {
	if !log.dropped {
		fmt.Printf("%s", "skipped")
		log.Drop()
	}

	return nil
}

func (log *Logger) Done() error {
	if !log.dropped {
		fmt.Printf("%s", "done")
		log.Drop()
	}

	return nil
}

func (log *Logger) Failed(err error) error {
	if !log.dropped {
		fmt.Printf("%s", "failed")
		log.Drop()
	}

	return err
}
