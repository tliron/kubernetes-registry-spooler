package common

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net/http"
)

func TLSTransport(certificatePath string, forceHttps bool) (http.RoundTripper, error) {
	if certPool, err := CertPool(certificatePath); err == nil {
		var transport http.RoundTripper
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		}
		if forceHttps {
			transport = NewForceHTTPSRoundTripper(transport)
		}
		return transport, nil
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

//
// ForceHTTPSRoundTripper
//

type ForceHTTPSRoundTripper struct {
	roundTripper http.RoundTripper
}

func NewForceHTTPSRoundTripper(roundTripper http.RoundTripper) *ForceHTTPSRoundTripper {
	return &ForceHTTPSRoundTripper{roundTripper}
}

// http.RoundTripper interface
func (self *ForceHTTPSRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Scheme != "https" {
		// Rewrite URL
		url := *request.URL
		url.Scheme = "https"
		request.URL = &url
	}

	return self.roundTripper.RoundTrip(request)
}
