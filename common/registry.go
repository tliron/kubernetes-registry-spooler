package common

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	namepkg "github.com/google/go-containerregistry/pkg/name"
	containerregistrypkg "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/stream"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	gzip "github.com/klauspost/pgzip"
)

func PushLayerToRegistry(readCloser io.ReadCloser, name string, transport http.RoundTripper) error {
	if tag, err := namepkg.NewTag(name); err == nil {
		// See: https://github.com/google/go-containerregistry/issues/707
		layer := stream.NewLayer(ioutil.NopCloser(readCloser))
		//layer = stream.NewLayer(readCloser)

		if image, err := mutate.AppendLayers(empty.Image, layer); err == nil {
			if transport != nil {
				return remote.Write(tag, image, remote.WithTransport(transport))
			} else {
				return remote.Write(tag, image)
			}
		} else {
			return err
		}
	} else {
		return err
	}
}

func PushTarballToRegistry(path string, name string, transport http.RoundTripper) error {
	if tag, err := namepkg.NewTag(name); err == nil {
		if image, err := tarball.ImageFromPath(path, &tag); err == nil {
			if transport != nil {
				return remote.Write(tag, image, remote.WithTransport(transport))
			} else {
				return remote.Write(tag, image)
			}
		} else {
			return err
		}
	} else {
		return err
	}
}

func PushGzippedTarballToRegistry(path string, name string, transport http.RoundTripper) error {
	if tag, err := namepkg.NewTag(name); err == nil {
		opener := func() (io.ReadCloser, error) {
			if reader, err := os.Open(path); err == nil {
				return gzip.NewReader(reader)
			} else {
				return nil, err
			}
		}

		if image, err := tarball.Image(opener, &tag); err == nil {
			if transport != nil {
				return remote.Write(tag, image, remote.WithTransport(transport))
			} else {
				return remote.Write(tag, image)
			}
		} else {
			return err
		}
	} else {
		return err
	}
}

func DeleteFromRegistry(name string, transport http.RoundTripper) error {
	if tag, err := namepkg.NewTag(name); err == nil {
		if image, err := remote.Image(tag); err == nil {
			if hash, err := image.Digest(); err == nil {
				digest := tag.Digest(hash.String())

				if transport != nil {
					return remote.Delete(digest, remote.WithTransport(transport))
				} else {
					return remote.Delete(digest)
				}
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

func PullTarballFromRegistry(name string, path string, transport http.RoundTripper) error {
	if tag, err := namepkg.NewTag(name); err == nil {
		var image containerregistrypkg.Image
		if transport != nil {
			image, err = remote.Image(tag, remote.WithTransport(transport))
		} else {
			image, err = remote.Image(tag)
		}

		if err == nil {
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

func ListImages(registry string, transport http.RoundTripper) ([]string, error) {
	if registry_, err := namepkg.NewInsecureRegistry(registry); err == nil {
		if transport != nil {
			return remote.Catalog(context.TODO(), registry_, remote.WithTransport(transport))
		} else {
			return remote.Catalog(context.TODO(), registry_)
		}
	} else {
		return nil, err
	}
}
