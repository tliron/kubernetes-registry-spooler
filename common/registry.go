package common

import (
	"io"
	"io/ioutil"
	"os"

	namepkg "github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/stream"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

func PushLayerToRegistry(readCloser io.ReadCloser, name string) error {
	if tag, err := namepkg.NewTag(name); err == nil {
		// See: https://github.com/google/go-containerregistry/issues/707
		layer := stream.NewLayer(ioutil.NopCloser(readCloser))
		//layer = stream.NewLayer(readCloser)

		if image, err := mutate.AppendLayers(empty.Image, layer); err == nil {
			return remote.Write(tag, image)
		} else {
			return err
		}
	} else {
		return err
	}
}

func PushTarballToRegistry(path string, name string) error {
	if tag, err := namepkg.NewTag(name); err == nil {
		if image, err := tarball.ImageFromPath(path, &tag); err == nil {
			return remote.Write(tag, image)
		} else {
			return err
		}
	} else {
		return err
	}
}

func DeleteFromRegistry(name string) error {
	if tag, err := namepkg.NewTag(name); err == nil {
		if image, err := remote.Image(tag); err == nil {
			if hash, err := image.Digest(); err == nil {
				digest := tag.Digest(hash.String())
				return remote.Delete(digest)
			} else {
				return err
			}
		} else {
			return err
		}
	} else {
		return err
	}
}

func PullTarballFromRegistry(name string, path string) error {
	if tag, err := namepkg.NewTag(name); err == nil {
		if image, err := remote.Image(tag); err == nil {
			var writer io.Writer
			if path == "" {
				writer = os.Stdout
			} else {
				if file, err := os.Create(path); err == nil {
					defer file.Close()
					writer = file
				} else {
					return err
				}
			}

			return tarball.Write(tag, image, writer)
		} else {
			return err
		}
	} else {
		return err
	}
}
