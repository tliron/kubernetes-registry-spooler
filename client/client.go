package client

import (
	"bytes"
	"io"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	Kubernetes           *kubernetes.Clientset
	REST                 rest.Interface
	Config               *rest.Config
	Namespace            string
	SpoolerAppName       string
	SpoolerContainerName string
	SpoolDirectory       string
}

func NewClient(kubernetes *kubernetes.Clientset, rest rest.Interface, config *rest.Config, namespace string, spoolerAppName string, spoolerContainerName string, spoolDirectory string) *Client {
	return &Client{
		Kubernetes:           kubernetes,
		REST:                 rest,
		Config:               config,
		Namespace:            namespace,
		SpoolerAppName:       spoolerAppName,
		SpoolerContainerName: spoolerContainerName,
		SpoolDirectory:       spoolDirectory,
	}
}

func (self *Client) Push(imageName string, reader io.Reader) error {
	if podName, err := self.getFirstPodName(); err == nil {
		path := self.getPath(imageName)
		tempPath := path + "~"
		if err := self.WriteToContainer(podName, reader, tempPath); err == nil {
			return self.Exec(podName, nil, nil, "mv", tempPath, path)
		} else {
			return err
		}
	} else {
		return err
	}
}

func (self *Client) Delete(imageName string) error {
	if podName, err := self.getFirstPodName(); err == nil {
		path := self.getPath(imageName) + "!"
		return self.Exec(podName, nil, nil, "touch", path)
	} else {
		return err
	}
}

func (self *Client) PullTarball(imageName string, writer io.Writer) error {
	if podName, err := self.getFirstPodName(); err == nil {
		return self.Exec(podName, nil, writer, "registry", "pull", imageName)
	} else {
		return err
	}
}

func (self *Client) List() ([]string, error) {
	if podName, err := self.getFirstPodName(); err == nil {
		var buffer bytes.Buffer
		if err := self.Exec(podName, nil, &buffer, "registry", "list"); err == nil {
			return strings.Split(strings.TrimRight(buffer.String(), "\n"), "\n"), nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}
