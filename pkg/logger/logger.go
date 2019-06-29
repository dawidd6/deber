package logger

import (
	"fmt"
	"strings"
)

const (
	cyan   = "\033[0;36m"
	blue   = "\033[0;34m"
	red    = "\033[0;31m"
	normal = "\033[0m"
)

type Logger struct {
	prefix string
	color  bool
}

func New(prefix string, color bool) *Logger {
	return &Logger{
		prefix: prefix,
		color:  color,
	}
}

func (log *Logger) Info(info ...string) {
	log.print("info", blue, strings.Join(info, " "))
}

func (log *Logger) Error(err error) {
	log.print("error", red, err)
}

func (log *Logger) Notice(notice ...string) {
	log.print("notice", cyan, strings.Join(notice, " "))
}

func (log *Logger) print(label, color string, v interface{}) {
	if log.color {
		fmt.Printf("%s%s:%s:%s %s\n", color, log.prefix, label, normal, v)
	} else {
		fmt.Printf("%s:%s: %s\n", log.prefix, label, v)
	}
}
