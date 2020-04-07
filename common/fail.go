package common

import (
	"fmt"
	"os"

	"github.com/tliron/kubernetes-registry-spooler/common/terminal"
)

func Fail(message string) {
	if !terminal.Quiet {
		fmt.Fprintln(terminal.Stderr, terminal.ColorError(message))
	}
	os.Exit(1)
}

func Failf(f string, args ...interface{}) {
	Fail(fmt.Sprintf(f, args...))
}

func FailOnError(err error) {
	if err != nil {
		Fail(err.Error())
	}
}
