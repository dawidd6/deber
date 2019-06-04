package logger

import (
	"errors"
	"fmt"
)

const (
	blue   = "\033[0;34m"
	red    = "\033[0;31m"
	normal = "\033[0m"
)

var Skip = errors.New("SKIP")

type Logger struct {
	Prefix  string
	Color   bool
	newLine bool
}

func New(prefix string) *Logger {
	return &Logger{
		Prefix:  prefix,
		Color:   true,
		newLine: true,
	}
}

func (log *Logger) Drop() {
	log.newLine = true
	fmt.Println()
}

func (log *Logger) Result(err error) {
	if log.newLine {
		return
	}

	if err == Skip {
		fmt.Printf("skipped\n")
	} else if err == nil {
		fmt.Printf("done\n")
	} else {
		fmt.Printf("failed\n")
	}
}

func (log *Logger) Info(v interface{}) {
	if log.Color {
		fmt.Printf("%s%s:info:%s %s ...", blue, log.Prefix, normal, v)
	} else {
		fmt.Printf("%s:info: %s ...", log.Prefix, v)
	}

	log.newLine = false
}

func (log *Logger) Error(v interface{}) {
	if log.Color {
		fmt.Printf("%s%s:error:%s %s\n", red, log.Prefix, normal, v)
	} else {
		fmt.Printf("%s:error: %s\n", log.Prefix, v)
	}
}
