package router

import (
	"crypto/tls"

	"Ustasjs/yp-url-shortener/internal/tlscert"
)

// generateTLSConfig builds a TLS configuration backed by a freshly generated
// self-signed certificate held entirely in memory, so HTTPS can be enabled
// without shipping certificate files alongside the binary.
func generateTLSConfig() (*tls.Config, error) {
	certPEM, keyPEM, err := tlscert.GeneratePEM()
	if err != nil {
		return nil, err
	}

	certificate, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tls.Config{Certificates: []tls.Certificate{certificate}}, nil
}

// tlsConfigFromFiles builds a TLS configuration from a certificate and key
// stored on disk, used when the operator supplies their own certificate.
func tlsConfigFromFiles(certFile, keyFile string) (*tls.Config, error) {
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	return &tls.Config{Certificates: []tls.Certificate{certificate}}, nil
}
