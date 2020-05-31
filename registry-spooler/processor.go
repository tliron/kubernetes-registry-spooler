package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/op/go-logging"
	"github.com/tliron/kubernetes-registry-spooler/common"
)

type Processor struct {
	registry string
	work     chan string
	log      *logging.Logger
}

func NewProcessor(registry string, queue int) *Processor {
	return &Processor{
		registry: registry,
		work:     make(chan string, queue),
		log:      logging.MustGetLogger("processor"),
	}
}

func (self *Processor) Enqueue(path string) {
	self.log.Debugf("enqueuing: %s", path)
	self.work <- path
}

func (self *Processor) Close() {
	close(self.work)
}

func (self *Processor) Run() {
	for self.Process() {
	}
}

func (self *Processor) Process() bool {
	if path, ok := <-self.work; ok {
		self.log.Debugf("processing: %s", path)

		if strings.HasSuffix(path, "!") {
			self.Delete(path)
		} else {
			self.Publish(path)
		}

		if err := os.Remove(path); err == nil {
			self.log.Infof("deleted file %s", path)
		} else {
			self.log.Errorf("could not delete file %s: %s", path, err.Error())
		}

		return true
	} else {
		self.log.Warning("no more work")
		return false
	}
}

func (self *Processor) Publish(path string) {
	name := self.getImageName(path)

	var err error
	if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".tgz") {
		self.log.Infof("publishing gzipped tarball %s to image %s", path, name)
		err = common.PushGzippedTarballToRegistry(path, name)
	} else if strings.HasSuffix(path, ".tar") {
		self.log.Infof("publishing tarball %s to image %s", path, name)
		err = common.PushTarballToRegistry(path, name)
	} else {
		self.log.Infof("publishing layer %s to image %s", path, name)
		if file, err2 := os.Open(path); err2 == nil {
			err = common.PushLayerToRegistry(file, name)
		} else {
			self.log.Errorf("could not read file %s: %s", path, err2.Error())
		}
	}

	if err == nil {
		self.log.Infof("published image %s", name)
	} else {
		self.log.Errorf("could not publish image %s: %s", name, err.Error())
	}
}

func (self *Processor) Delete(path string) {
	name := self.getImageName(path[:len(path)-1])

	self.log.Infof("deleting image %s", name)
	if err := common.DeleteFromRegistry(name); err == nil {
		self.log.Infof("deleted image %s", name)
	} else {
		self.log.Errorf("could not delete image %s: %s", name, err.Error())
	}
}

func (self *Processor) getImageName(path string) string {
	name := filepath.Base(path)
	if dot := strings.Index(name, "."); dot != -1 {
		// Note: filepath.Ext will return the last extension only
		name = name[:dot]
	}
	name = strings.ReplaceAll(name, "\\", "/")
	return fmt.Sprintf("%s/%s", self.registry, name)
}
