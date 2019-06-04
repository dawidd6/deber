package logger

import (
	"fmt"
)

const (
	blue   = "\033[0;34m"
	red    = "\033[0;31m"
	normal = "\033[0m"
)

type Logger struct {
	Prefix string
	Color  bool
}

func New(prefix string) *Logger {
	return &Logger{
		Prefix: prefix,
		Color:  true,
	}
}

func (log *Logger) Info(v interface{}) {
	if log.Color {
		fmt.Printf("%s%s:info:%s %s ...\n", blue, log.Prefix, normal, v)
	} else {
		fmt.Printf("%s:info: %s ...\n", log.Prefix, v)
	}
}

func (log *Logger) Error(v interface{}) {
	if log.Color {
		fmt.Printf("%s%s:error:%s %s\n", red, log.Prefix, normal, v)
	} else {
		fmt.Printf("%s:error: %s\n", log.Prefix, v)
	}
}
