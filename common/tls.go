package common

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net/http"

	"github.com/tliron/kutil/util"
)

func TLSRoundTripper(certificatePath string) (http.RoundTripper, error) {
	if certPool, err := CertPool(certificatePath); err == nil {
		// We need to force HTTPS because go-containerregistry will attempt to drop down to HTTP for local addresses
		return util.NewForceHTTPSRoundTripper(&http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		}), nil
	} else {
		return nil, err
	}
}

func CertPool(certificatePath string) (*x509.CertPool, error) {
	if certificate, err := Certificate(certificatePath); err == nil {
		certPool := x509.NewCertPool()
		certPool.AddCert(certificate)
		return certPool, nil
	} else {
		return nil, err
	}
}

func Certificate(certificatePath string) (*x509.Certificate, error) {
	if bytes, err := ioutil.ReadFile(certificatePath); err == nil {
		block, _ := pem.Decode(bytes)
		return x509.ParseCertificate(block.Bytes)
	} else {
		return nil, err
	}
}
