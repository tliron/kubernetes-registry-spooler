package main

import (
	"fmt"

	"github.com/tliron/kubernetes-registry-spooler/common"
)

func Pull(registry string, name string, path string) {
	name = fmt.Sprintf("%s/%s", registry, name)
	err := common.PullTarballFromRegistry(name, path)
	common.FailOnError(err)
}
