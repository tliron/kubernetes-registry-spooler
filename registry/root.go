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
	rootCommand.PersistentFlags().StringVarP(&logTo, "log", "l", "", "log to file (defaults to stderr)")
	rootCommand.PersistentFlags().CountVarP(&verbose, "verbose", "v", "add a log verbosity level (can be used twice)")
	rootCommand.PersistentFlags().BoolVarP(&colorize, "colorize", "z", true, "colorize output")
	rootCommand.PersistentFlags().StringVarP(&registry, "registry", "r", "localhost:5000", "registry URL")
}

var rootCommand = &cobra.Command{
	Use:   "registry [IMAGE NAME]",
	Short: "Access the container image registry",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if colorize {
			terminal.EnableColor()
		}
		if logTo == "" {
			common.ConfigureLogging(verbose, nil)
		} else {
			common.ConfigureLogging(verbose, &logTo)
		}
		if registry == "" {
			common.Fail("must provide \"--registry\"")
		}
	},
}
