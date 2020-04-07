package main

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kubernetes-registry-spooler/common"
	"github.com/tliron/kubernetes-registry-spooler/common/terminal"
)

var logTo string
var verbose int
var colorize bool
var registry string

func init() {
	command.PersistentFlags().StringVarP(&logTo, "log", "l", "", "log to file (defaults to stderr)")
	command.PersistentFlags().CountVarP(&verbose, "verbose", "v", "add a log verbosity level (can be used twice)")
	command.PersistentFlags().BoolVarP(&colorize, "colorize", "z", true, "colorize output")
	command.PersistentFlags().StringVarP(&registry, "registry", "r", "localhost:5000", "registry URL")
}

var command = &cobra.Command{
	Use:   "registry-pull [IMAGE NAME] [TAR FILE PATH]",
	Short: "Pull tar file from a container image registry",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if colorize {
			terminal.EnableColor()
		}
		if logTo == "" {
			common.ConfigureLogging(verbose, nil)
		} else {
			common.ConfigureLogging(verbose, &logTo)
		}
	},
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if registry == "" {
			common.Fail("must provide \"--registry\"")
		}

		name := args[0]
		path := args[1]

		Pull(registry, name, path)
	},
}
