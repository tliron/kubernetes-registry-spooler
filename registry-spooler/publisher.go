package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofrs/flock"
	"github.com/op/go-logging"
	"github.com/tliron/kubernetes-registry-spooler/common"
)

type Publisher struct {
	registry  string
	transport http.RoundTripper
	work      chan string
	log       *logging.Logger
}

func NewPublisher(registry string, queue int) *Publisher {
	log.Infof("certificate path: %s", certificatePath)
	log.Infof("force HTTPS: %t", forceHttps)

	transport, err := common.TLSTransport(certificatePath, forceHttps)
	if err != nil {
		log.Warningf("%s", err.Error())
	}

	return &Publisher{
		registry:  registry,
		transport: transport,
		work:      make(chan string, queue),
		log:       logging.MustGetLogger("publisher"),
	}
}

func (self *Publisher) Enqueue(path string) {
	self.log.Debugf("enqueuing: %s", path)
	self.work <- path
}

func (self *Publisher) Close() {
	close(self.work)
}

func (self *Publisher) Run() {
	defer self.Close()
	for self.Process() {
	}
}

func (self *Publisher) Process() bool {
	if path, ok := <-self.work; ok {
		// Lock file
		lock := flock.New(path)
		if err := lock.Lock(); err == nil {
			defer lock.Unlock()
		} else {
			self.log.Errorf("could not lock file %q: %s", path, err.Error())
			return true
		}

		// File may have already been deleted by another process
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				self.log.Infof("file %q already deleted", path)
			} else {
				self.log.Errorf("could not access file %q: %s", path, err.Error())
			}
			return true
		}

		/// Process
		if strings.HasSuffix(path, "!") {
			self.Delete(path[:len(path)-1])
		} else {
			self.Publish(path)
		}

		// Delete file
		if err := os.Remove(path); err == nil {
			self.log.Infof("deleted file %q", path)
		} else {
			self.log.Errorf("could not delete file %q: %s", path, err.Error())
		}

		return true
	} else {
		self.log.Warning("no more work")
		return false
	}
}

func (self *Publisher) Publish(path string) {
	name := self.getImageName(path)

	var err error
	if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".tgz") {
		self.log.Infof("publishing gzipped tarball %q to image %q", path, name)
		err = common.PushGzippedTarballToRegistry(path, name, self.transport)
	} else if strings.HasSuffix(path, ".tar") {
		self.log.Infof("publishing tarball %q to image %q", path, name)
		err = common.PushTarballToRegistry(path, name, self.transport)
	} else {
		self.log.Infof("publishing layer %q to image %q", path, name)
		if file, err2 := os.Open(path); err2 == nil {
			err = common.PushLayerToRegistry(file, name, self.transport)
		} else {
			self.log.Errorf("could not read file %q: %s", path, err2.Error())
		}
	}

	if err == nil {
		self.log.Infof("published image %q", name)
	} else {
		self.log.Errorf("could not publish image %q: %s", name, err.Error())
	}
}

func (self *Publisher) Delete(path string) {
	name := self.getImageName(path)
	self.log.Infof("deleting image %q", name)
	if err := common.DeleteFromRegistry(name, self.transport); err == nil {
		self.log.Infof("deleted image %q", name)
	} else {
		self.log.Errorf("could not delete image %q: %s", name, err.Error())
	}
}

func (self *Publisher) getImageName(path string) string {
	name := filepath.Base(path)
	if dot := strings.Index(name, "."); dot != -1 {
		// Note: filepath.Ext will return the last extension only
		name = name[:dot]
	}
	name = strings.ReplaceAll(name, "\\", "/")
	return fmt.Sprintf("%s/%s", self.registry, name)
}
