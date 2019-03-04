package main

import (
	"fmt"
	"time"
)

type Loading struct {
	stop    chan bool
	current int
	dots    int
	delay   time.Duration
}

func NewLoading() *Loading {
	return &Loading{
		stop:    make(chan bool, 1),
		current: 1,
		dots:    5,
		delay:   time.Second / 4,
	}
}

func (loading *Loading) Reset() {
	loading.current = 1
}

func (loading *Loading) Start() {
	go func() {
		for {
			select {
			case <-time.After(loading.delay):
				line := make([]rune, loading.dots)
				for i := 0; i < loading.dots; i++ {
					line[i] = ' '
				}
				for i := 0; i < loading.current; i++ {
					line[i] = '.'
				}

				fmt.Printf("\r%s", string(line))

				if loading.current == loading.dots {
					loading.current = 0
				} else {
					loading.current++
				}
			case <-loading.stop:
				return
			}
		}
	}()
}

func (loading *Loading) Stop() {
	loading.stop <- true

	line := make([]rune, loading.dots)
	for i := 0; i < loading.dots; i++ {
		line[i] = '.'
	}

	fmt.Printf("\r%s\n", string(line))
}
