package client

import (
	"io"
	"path/filepath"

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
	PullDirectory        string
}

func NewClient(kubernetes *kubernetes.Clientset, rest rest.Interface, config *rest.Config, namespace string, spoolerAppName string, spoolerContainerName string, spoolDirectory string, pullDirectory string) *Client {
	return &Client{
		Kubernetes:           kubernetes,
		REST:                 rest,
		Config:               config,
		Namespace:            namespace,
		SpoolerAppName:       spoolerAppName,
		SpoolerContainerName: spoolerContainerName,
		SpoolDirectory:       spoolDirectory,
		PullDirectory:        pullDirectory,
	}
}

func (self *Client) Push(imageName string, reader io.Reader) error {
	if podName, err := self.getFirstPodName(); err == nil {
		path := filepath.Join(self.SpoolDirectory, imageName)
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
		path := filepath.Join(self.SpoolDirectory, imageName) + "!"
		return self.Exec(podName, nil, nil, "touch", path)
	} else {
		return err
	}
}

func (self *Client) PullTarball(imageName string, writer io.Writer) error {
	if podName, err := self.getFirstPodName(); err == nil {
		remotePath := filepath.Join(self.PullDirectory, imageName)
		if err := self.Exec(podName, nil, nil, "registry-pull", imageName, remotePath); err == nil {
			err := self.ReadFromContainer(podName, writer, remotePath)
			self.Exec(podName, nil, nil, "rm", remotePath)
			return err
		} else {
			return err
		}
	} else {
		return err
	}
}
