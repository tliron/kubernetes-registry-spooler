package common

import (
	"io"
	"io/ioutil"

	namepkg "github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/stream"
)

func PushToRegistry(readCloser io.ReadCloser, name string) error {
	tag, err := namepkg.NewTag(name)
	if err != nil {
		return err
	}

	// See: https://github.com/google/go-containerregistry/issues/707
	layer := stream.NewLayer(ioutil.NopCloser(readCloser))
	//layer = stream.NewLayer(readCloser)

	image, err := mutate.AppendLayers(empty.Image, layer)
	if err != nil {
		return err
	}

	return remote.Write(tag, image)
}

func DeleteFromRegistry(name string) error {
	tag, err := namepkg.NewTag(name)
	if err != nil {
		return err
	}

	image, err := remote.Image(tag)
	if err != nil {
		return err
	}

	hash, err := image.Digest()
	if err != nil {
		return err
	}

	digest := tag.Digest(hash.String())

	return remote.Delete(digest)
}
