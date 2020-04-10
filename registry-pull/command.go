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
var output string

func init() {
	command.PersistentFlags().StringVarP(&logTo, "log", "l", "", "log to file (defaults to stderr)")
	command.PersistentFlags().CountVarP(&verbose, "verbose", "v", "add a log verbosity level (can be used twice)")
	command.PersistentFlags().BoolVarP(&colorize, "colorize", "z", true, "colorize output")
	command.PersistentFlags().StringVarP(&registry, "registry", "r", "localhost:5000", "registry URL")
	command.PersistentFlags().StringVarP(&output, "output", "o", "", "output to file (defaults to stdout)")
}

var command = &cobra.Command{
	Use:   "registry-pull [IMAGE NAME]",
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
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if registry == "" {
			common.Fail("must provide \"--registry\"")
		}

		name := args[0]

		Pull(registry, name, output)
	},
}
