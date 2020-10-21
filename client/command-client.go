package client

import (
	"bytes"
	"io"
	"strings"
)

//
// CommandClient
//

type CommandClient struct {
	Client              *Client
	Registry            string
	RegistryCertificate string
}

func NewCommandClient(client *Client, registry string, registryCertificate string) *CommandClient {
	if registry == "" {
		// Default for sidecars
		registry = "localhost:5000"
	}

	return &CommandClient{
		Client:              client,
		Registry:            registry,
		RegistryCertificate: registryCertificate,
	}
}

func (self *CommandClient) PullTarball(imageName string, writer io.Writer) error {
	return self.RegistryClient(writer, "pull", imageName)
}

func (self *CommandClient) List() ([]string, error) {
	var buffer bytes.Buffer
	if err := self.RegistryClient(&buffer, "list"); err == nil {
		return strings.Split(strings.TrimRight(buffer.String(), "\n"), "\n"), nil
	} else {
		return nil, err
	}
}

// Utils

func (self *CommandClient) RegistryClient(writer io.Writer, arguments ...string) error {
	if podName, err := self.Client.getFirstPodName(); err == nil {
		arguments = append([]string{"registry-client"}, arguments...)
		arguments = append(arguments, "--registry", self.Registry)
		if self.RegistryCertificate != "" {
			arguments = append(arguments, "--certificate", self.RegistryCertificate)
		}

		return self.Client.Exec(podName, nil, writer, arguments...)
	} else {
		return err
	}
}
