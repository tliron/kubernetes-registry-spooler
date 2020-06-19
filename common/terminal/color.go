package terminal

import (
	"github.com/zchee/color/v2"
)

var colorize = false

func EnableColor() {
	colorize = true
	Stdout = color.Output
	Stderr = color.Error
}

type Colorizer = func(name string) string

func ColorError(name string) string {
	if colorize {
		return color.RedString(name)
	} else {
		return name
	}
}
