package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tliron/kubernetes-registry-spooler/common"
)

func init() {
	rootCommand.AddCommand(listCommand)
}

var listCommand = &cobra.Command{
	Use:   "list",
	Short: "List images a container image registry",
	Run: func(cmd *cobra.Command, args []string) {
		List(registry)
	},
}

func List(registry string) {
	images, err := common.ListImages(registry)
	common.FailOnError(err)
	for _, image := range images {
		fmt.Println(image)
	}
}
