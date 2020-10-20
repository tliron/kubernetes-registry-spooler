package main

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/tliron/kubernetes-registry-spooler/common"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

var logTo string
var verbose int
var colorize string
var registry string
var certificatePath string
var forceHttps bool

var transport http.RoundTripper

func init() {
	rootCommand.PersistentFlags().StringVarP(&logTo, "log", "l", "", "log to file (defaults to stderr)")
	rootCommand.PersistentFlags().CountVarP(&verbose, "verbose", "v", "add a log verbosity level (can be used twice)")
	rootCommand.PersistentFlags().StringVarP(&colorize, "colorize", "z", "true", "colorize output (boolean or \"force\")")
	rootCommand.PersistentFlags().StringVarP(&registry, "registry", "r", "localhost:5000", "registry URL")
	rootCommand.PersistentFlags().StringVarP(&certificatePath, "certificate", "c", "/secret/tls.crt", "registry TLS certificate file path (in PEM format)")
	rootCommand.PersistentFlags().BoolVarP(&forceHttps, "force-https", "s", false, "force HTTPS connections to registry (HTTP is used by default for local addresses)")
}

var rootCommand = &cobra.Command{
	Use:   "registry",
	Short: "Access the container image registry",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := terminal.ProcessColorizeFlag(colorize)
		util.FailOnError(err)
		if logTo == "" {
			util.ConfigureLogging(verbose, nil)
		} else {
			util.ConfigureLogging(verbose, &logTo)
		}
		if registry == "" {
			util.Fail("must provide \"--registry\"")
		}
		transport, err = common.TLSTransport(certificatePath, forceHttps)
		if err != nil {
			fmt.Fprintf(terminal.Stderr, "%s\n", err.Error())
		}
	},
}
