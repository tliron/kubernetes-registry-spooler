package client

import (
	"context"
	"fmt"
	"path/filepath"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (self *Client) getPath(imageName string) string {
	return filepath.Join(self.SpoolDirectory, imageName)
}

func (self *Client) getPods() (*core.PodList, error) {
	labels_ := labels.Set(map[string]string{
		"app.kubernetes.io/name": self.SpoolerAppName,
	})
	selector := labels_.AsSelector().String()

	return self.Kubernetes.CoreV1().Pods(self.Namespace).List(context.TODO(), meta.ListOptions{LabelSelector: selector})
}

func (self *Client) getPodNames() ([]string, error) {
	if pods, err := self.getPods(); err == nil {
		length := len(pods.Items)
		if length == 0 {
			return nil, fmt.Errorf("no pods for app.kubernetes.io/name=\"%s\" in namespace \"%s\"", self.SpoolerAppName, self.Namespace)
		}

		names := make([]string, length)
		for index, pod := range pods.Items {
			names[index] = pod.Name
		}

		return names, nil
	} else {
		return nil, err
	}
}

func (self *Client) getFirstPodName() (string, error) {
	if names, err := self.getPodNames(); err == nil {
		return names[0], nil
	} else {
		return "", err
	}
}
