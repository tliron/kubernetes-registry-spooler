package main

import (
	"github.com/tliron/kubernetes-registry-spooler/common"
)

func main() {
	err := rootCommand.Execute()
	common.FailOnError(err)
}
