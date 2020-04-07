package main

import (
	"github.com/tliron/kubernetes-registry-spooler/common"
)

func main() {
	err := command.Execute()
	common.FailOnError(err)
}
