// Package tlscert generates self-signed TLS certificates. It is the single
// source of the certificate-generation logic shared by the HTTPS server (which
// keeps the certificate in memory) and the cmd/certgen utility (which writes it
// to files).
package tlscert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

// GeneratePEM creates a 2048-bit RSA key and a self-signed certificate for it,
// and returns them PEM-encoded. The certificate is valid for the loopback
// addresses (127.0.0.1 and ::1) for ten years, which is enough for local HTTPS
// during development.
func GeneratePEM() (certPEM, keyPEM []byte, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("generate private key: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1454),
		Subject: pkix.Name{
			Organization: []string{"Ustasjs/yp-url-shortener"},
			Country:      []string{"RU"},
		},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("create certificate: %w", err)
	}

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	return certPEM, keyPEM, nil
}
