package common

import (
	"os"

	"github.com/op/go-logging"
	"github.com/tliron/kubernetes-registry-spooler/common/terminal"
)

var plainFormatter = logging.MustStringFormatter(
	`%{time:2006/01/02 15:04:05.000} %{level:8.8s} [%{module}] %{message}`,
)

var colorFormatter = logging.MustStringFormatter(
	`%{color}%{time:2006/01/02 15:04:05.000} %{level:8.8s} [%{module}] %{message}%{color:reset}`,
)

const logFileWritePermissions = 0600

func ConfigureLogging(verbosity int, path *string) {
	var backend *logging.LogBackend
	if path != nil {
		if file, err := os.OpenFile(*path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, logFileWritePermissions); err == nil {
			// defer f.Close() ???
			backend = logging.NewLogBackend(file, "", 0)
			logging.SetFormatter(plainFormatter)
		} else {
			Failf("log file error: %s", err.Error())
		}
	} else {
		backend = logging.NewLogBackend(terminal.Stderr, "", 0)
		logging.SetFormatter(colorFormatter)
	}

	leveledBackend := logging.AddModuleLevel(backend)

	verbosity += 3 // 0 verbosity is NOTICE
	if verbosity > 5 {
		verbosity = 5
	}
	level := logging.Level(verbosity)

	leveledBackend.SetLevel(level, "")

	logging.SetBackend(leveledBackend)
}
