package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/heptiolabs/healthcheck"
	"github.com/op/go-logging"
	"github.com/tliron/kutil/util"
)

var log = logging.MustGetLogger("registry-spooler")

func RunSpooler(registryUrl string, path string) {
	stopChannel := util.SetupSignalHandler()

	processor := NewPublisher(registryUrl, queue)
	log.Info("starting processor")
	go processor.Run()
	defer processor.Close()

	fileInfos, err := ioutil.ReadDir(path)
	util.FailOnError(err)
	for _, fileInfo := range fileInfos {
		processor.Enqueue(filepath.Join(path, fileInfo.Name()))
	}

	watcher, err := NewWatcher()
	util.FailOnError(err)

	err = watcher.Add(path, fsnotify.Create, func(path string) {
		processor.Enqueue(path)
	})
	util.FailOnError(err)

	log.Info("starting watcher")
	go watcher.Run()

	go func() {
		log.Info("starting health monitor")
		health := healthcheck.NewHandler()
		err := http.ListenAndServe(fmt.Sprintf(":%d", healthPort), health)
		util.FailOnError(err)
	}()

	<-stopChannel
}
