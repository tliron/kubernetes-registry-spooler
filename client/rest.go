package client

import (
	"io"

	kubernetesutil "github.com/tliron/kutil/kubernetes"
)

func (self *Client) Write(podName string, reader io.Reader, targetPath string) error {
	return kubernetesutil.WriteToContainer(self.REST, self.Config, self.Namespace, podName, self.SpoolerContainerName, reader, targetPath, nil)
}

func (self *Client) Read(podName string, writer io.Writer, sourcePath string) error {
	return kubernetesutil.ReadFromContainer(self.REST, self.Config, self.Namespace, podName, self.SpoolerContainerName, writer, sourcePath)
}

func (self *Client) Move(podName string, fromPath string, toPath string) error {
	return self.Exec(podName, nil, nil, "mv", fromPath, toPath)
}

func (self *Client) Touch(podName string, path string) error {
	return self.Exec(podName, nil, nil, "touch", path)
}

func (self *Client) Exec(podName string, stdin io.Reader, stdout io.Writer, command ...string) error {
	return kubernetesutil.Exec(self.REST, self.Config, self.Namespace, podName, self.SpoolerContainerName, stdin, stdout, self.Stderr, false, command...)
}
