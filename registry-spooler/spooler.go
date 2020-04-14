package main

import (
	"io/ioutil"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/tliron/kubernetes-registry-spooler/common"
)

func Spool(registryUrl string, path string) {
	stopChannel := common.SetupSignalHandler()

	processor := NewProcessor(registryUrl, queue)
	go processor.Run()
	defer processor.Close()

	fileInfos, err := ioutil.ReadDir(path)
	common.FailOnError(err)
	for _, fileInfo := range fileInfos {
		processor.Enqueue(filepath.Join(path, fileInfo.Name()))
	}

	watcher, err := NewWatcher()
	common.FailOnError(err)

	err = watcher.Add(path, fsnotify.Create, func(path string) {
		processor.Enqueue(path)
	})
	common.FailOnError(err)
	go watcher.Run()

	<-stopChannel
}
