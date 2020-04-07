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

func (self *Processor) Add(path string) {
	self.log.Debugf("queuing: %s", path)
	self.work <- path
}

func (self *Processor) Close() {
	close(self.work)
}

func (self *Processor) Run() {
	defer self.Close()
	for self.Process() {
	}
}

func (self *Processor) Process() bool {
	if path, ok := <-self.work; ok {
		self.log.Debugf("processing: %s", path)
		if strings.HasSuffix(path, "!") {
			self.Delete(path)
		} else {
			self.Push(path)
		}
		return true
	} else {
		self.log.Warning("no more work")
		return false
	}
}

func (self *Processor) Push(path string) {
	if file, err := os.Open(path); err == nil {
		name := self.getImageName(path)

		self.log.Infof("pushing file %s to image %s", path, name)
		if err := common.PushToRegistry(file, name); err == nil {
			self.log.Infof("pushed image %s", name)
			if err := os.Remove(path); err == nil {
				self.log.Infof("deleted file %s", path)
			} else {
				self.log.Errorf("could not delete file %s: %s", path, err.Error())
			}
		} else {
			self.log.Errorf("could not push image %s: %s", name, err.Error())
		}
	} else {
		self.log.Errorf("could not read file %s: %s", path, err.Error())
	}
}

func (self *Processor) Delete(path string) {
	name := self.getImageName(path[:len(path)-1])

	self.log.Infof("deleting image %s", name)
	if err := common.DeleteFromRegistry(name); err == nil {
		self.log.Infof("deleted image %s", name)
		if err := os.Remove(path); err == nil {
			self.log.Infof("deleted file %s", path)
		} else {
			self.log.Errorf("could not delete file %s: %s", path, err.Error())
		}
	} else {
		self.log.Errorf("could not delete image %s: %s", name, err.Error())
	}
}

func (self *Processor) getImageName(path string) string {
	extension := filepath.Ext(path)
	name := filepath.Base(path)
	name = name[:len(name)-len(extension)]
	return fmt.Sprintf("%s/%s", self.registry, name)
}
