package client

import (
	contextpkg "context"
	"io"
	"path/filepath"
	"strings"

	kubernetesutil "github.com/tliron/kutil/kubernetes"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

//
// Client
//

type Client struct {
	Kubernetes kubernetes.Interface
	REST       rest.Interface
	Config     *rest.Config
	Context    contextpkg.Context
	Stderr     io.Writer

	Namespace            string
	SpoolerAppName       string
	SpoolerContainerName string
	SpoolDirectory       string
}

func NewClient(kubernetes kubernetes.Interface, rest rest.Interface, config *rest.Config, context contextpkg.Context, stderr io.Writer, namespace string, spoolerAppName string, spoolerContainerName string, spoolDirectory string) *Client {
	return &Client{
		Kubernetes: kubernetes,
		REST:       rest,
		Config:     config,
		Context:    contextpkg.TODO(),
		Stderr:     stderr,

		Namespace:            namespace,
		SpoolerAppName:       spoolerAppName,
		SpoolerContainerName: spoolerContainerName,
		SpoolDirectory:       spoolDirectory,
	}
}

func (self *Client) Publish(imageName string, reader io.Reader) error {
	if podName, err := self.getFirstPodName(); err == nil {
		path := self.getPath(imageName)
		tempPath := path + "~"
		if err := self.Write(podName, reader, tempPath); err == nil {
			return self.Move(podName, tempPath, path)
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
		return self.Touch(podName, path)
	} else {
		return err
	}
}

// Utils

func (self *Client) getFirstPodName() (string, error) {
	return kubernetesutil.GetFirstPodName(self.Context, self.Kubernetes, self.Namespace, self.SpoolerAppName)
}

func (self *Client) getPath(imageName string) string {
	return filepath.Join(self.SpoolDirectory, strings.ReplaceAll(imageName, "/", "\\"))
}
