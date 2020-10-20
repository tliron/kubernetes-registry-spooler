package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tliron/kubernetes-registry-spooler/common"
	"github.com/tliron/kutil/util"
)

var output string

func init() {
	rootCommand.AddCommand(pullCommand)
	pullCommand.PersistentFlags().StringVarP(&output, "output", "o", "", "output to file (defaults to stdout)")
}

var pullCommand = &cobra.Command{
	Use:   "pull [IMAGE NAME]",
	Short: "Pull tarball from a container image registry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		Pull(registry, name, output)
	},
}

func Pull(registry string, name string, path string) {
	name = fmt.Sprintf("%s/%s", registry, name)
	err := common.PullTarballFromRegistry(name, path, transport)
	util.FailOnError(err)
}
