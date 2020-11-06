package common

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	namepkg "github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/stream"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	gzip "github.com/klauspost/pgzip"
)

type Client struct {
	secure  bool
	options []remote.Option
}

func NewClient(transport http.RoundTripper, username string, password string, token string) *Client {
	var options []remote.Option

	if transport != nil {
		options = append(options, remote.WithTransport(transport))
	}

	if (username != "") || (token != "") {
		authenticator := authn.FromConfig(authn.AuthConfig{
			Username:      username,
			Password:      password,
			RegistryToken: token,
		})
		options = append(options, remote.WithAuth(authenticator))
	}

	return &Client{
		secure:  transport != nil,
		options: options,
	}
}

func (self *Client) PushLayerToRegistry(readCloser io.ReadCloser, name string) error {
	if tag, err := namepkg.NewTag(name); err == nil {
		// See: https://github.com/google/go-containerregistry/issues/707
		layer := stream.NewLayer(ioutil.NopCloser(readCloser))
		//layer = stream.NewLayer(readCloser)

		if image, err := mutate.AppendLayers(empty.Image, layer); err == nil {
			return remote.Write(tag, image, self.options...)
		} else {
			return err
		}
	} else {
		return err
	}
}

func (self *Client) PushTarballToRegistry(path string, name string) error {
	if tag, err := namepkg.NewTag(name); err == nil {
		if image, err := tarball.ImageFromPath(path, &tag); err == nil {
			return remote.Write(tag, image, self.options...)
		} else {
			return err
		}
	} else {
		return err
	}
}

func (self *Client) PushGzippedTarballToRegistry(path string, name string) error {
	if tag, err := namepkg.NewTag(name); err == nil {
		opener := func() (io.ReadCloser, error) {
			if reader, err := os.Open(path); err == nil {
				return gzip.NewReader(reader)
			} else {
				return nil, err
			}
		}

		if image, err := tarball.Image(opener, &tag); err == nil {
			return remote.Write(tag, image, self.options...)
		} else {
			return err
		}
	} else {
		return err
	}
}

func (self *Client) DeleteFromRegistry(name string) error {
	if tag, err := namepkg.NewTag(name); err == nil {
		if image, err := remote.Image(tag, self.options...); err == nil {
			if hash, err := image.Digest(); err == nil {
				digest := tag.Digest(hash.String())

				return remote.Delete(digest, self.options...)
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

func (self *Client) PullTarballFromRegistry(name string, path string) error {
	if tag, err := namepkg.NewTag(name); err == nil {
		if image, err := remote.Image(tag, self.options...); err == nil {
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

func (self *Client) ListImages(registry string) ([]string, error) {
	if registry_, err := self.newRegistry(registry); err == nil {
		return remote.Catalog(context.TODO(), registry_, self.options...)
	} else {
		return nil, err
	}
}

// Utils

func (self *Client) newRegistry(registry string) (namepkg.Registry, error) {
	if self.secure {
		return namepkg.NewRegistry(registry)
	} else {
		return namepkg.NewRegistry(registry, namepkg.Insecure)
	}
}
