package main

import (
	"github.com/spf13/cobra"
	"github.com/tliron/kubernetes-registry-spooler/common"
	"github.com/tliron/kubernetes-registry-spooler/common/terminal"
)

var logTo string
var verbose int
var colorize bool
var directoryPath string
var registry string
var queue int

func init() {
	command.PersistentFlags().StringVarP(&logTo, "log", "l", "", "log to file (defaults to stderr)")
	command.PersistentFlags().CountVarP(&verbose, "verbose", "v", "add a log verbosity level (can be used twice)")
	command.PersistentFlags().BoolVarP(&colorize, "colorize", "z", true, "colorize output")
	command.PersistentFlags().StringVarP(&directoryPath, "directory", "d", "/spool", "spool directory path")
	command.PersistentFlags().StringVarP(&registry, "registry", "r", "localhost:5000", "registry URL")
	command.PersistentFlags().IntVarP(&queue, "queue", "q", 10, "maximum number of files to queue at once")
	common.SetCobraFlagsFromEnvironment("REGISTRY_SPOOLER_", command)
}

var command = &cobra.Command{
	Use:   "registry-spooler",
	Short: "Spooler for a container image registry",
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
	Run: func(cmd *cobra.Command, args []string) {
		if (directoryPath == "") || (registry == "") {
			common.Fail("must provide \"--directory\" and \"--registry\"")
		}

		Spool(registry, directoryPath)
	},
}
