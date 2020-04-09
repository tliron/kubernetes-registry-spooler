package client

import (
	"io"
	"os"
	"path/filepath"

	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

func (self *Client) CopyToContainer(podName string, sourcePath string, targetPath string) error {
	if reader, err := os.Open(sourcePath); err == nil {
		defer reader.Close()
		return self.WriteToContainer(podName, reader, targetPath)
	} else {
		return err
	}
}

func (self *Client) CopyFromContainer(podName string, sourcePath string, targetPath string) error {
	if writer, err := os.Create(targetPath); err == nil {
		defer writer.Close()
		return self.ReadFromContainer(podName, writer, sourcePath)
	} else {
		return err
	}
}

func (self *Client) WriteToContainer(podName string, reader io.Reader, targetPath string) error {
	dir := filepath.Dir(targetPath)
	if err := self.Exec(podName, nil, nil, "mkdir", "--parents", dir); err != nil {
		return err
	}

	// See: https://stackoverflow.com/a/57952887
	return self.Exec(podName, reader, nil, "cp", "/dev/stdin", targetPath)
}

func (self *Client) ReadFromContainer(podName string, writer io.Writer, sourcePath string) error {
	return self.Exec(podName, nil, writer, "cat", sourcePath)
}

func (self *Client) Exec(podName string, stdin io.Reader, stdout io.Writer, command ...string) error {
	execOptions := core.PodExecOptions{
		Container: self.SpoolerContainerName,
		Command:   command,
		Stderr:    true,
		TTY:       false,
	}

	streamOptions := remotecommand.StreamOptions{
		Stderr: os.Stderr,
		Tty:    false,
	}

	if stdin != nil {
		execOptions.Stdin = true
		streamOptions.Stdin = stdin
	}

	if stdout != nil {
		execOptions.Stdout = true
		streamOptions.Stdout = stdout
	}

	request := self.REST.Post().Namespace(self.Namespace).Resource("pods").Name(podName).SubResource("exec").VersionedParams(&execOptions, scheme.ParameterCodec)

	if executor, err := remotecommand.NewSPDYExecutor(self.Config, "POST", request.URL()); err == nil {
		if err = executor.Stream(streamOptions); err == nil {
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}
