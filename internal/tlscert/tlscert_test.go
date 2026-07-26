package tlscert

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net"
	"testing"
	"time"
)

func TestGeneratePEMLoadsAsKeyPair(t *testing.T) {
	certPEM, keyPEM, err := GeneratePEM()
	if err != nil {
		t.Fatalf("GeneratePEM: %v", err)
	}

	if _, err := tls.X509KeyPair(certPEM, keyPEM); err != nil {
		t.Fatalf("X509KeyPair: %v", err)
	}
}

func TestGeneratePEMCertificateProperties(t *testing.T) {
	certPEM, _, err := GeneratePEM()
	if err != nil {
		t.Fatalf("GeneratePEM: %v", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		t.Fatalf("cert PEM block = %v, want CERTIFICATE", block)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("ParseCertificate: %v", err)
	}

	now := time.Now()
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		t.Errorf("certificate not currently valid: [%v, %v]", cert.NotBefore, cert.NotAfter)
	}

	wantIPs := []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
	for _, want := range wantIPs {
		found := false
		for _, got := range cert.IPAddresses {
			if got.Equal(want) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing loopback SAN %v", want)
		}
	}

	// Self-signed: issuer equals subject and it verifies against itself.
	pool := x509.NewCertPool()
	pool.AddCert(cert)
	if _, err := cert.Verify(x509.VerifyOptions{Roots: pool, KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}); err != nil {
		t.Errorf("self-signed verify failed: %v", err)
	}
}

func TestGeneratePEMIsRandomPerCall(t *testing.T) {
	_, key1, err := GeneratePEM()
	if err != nil {
		t.Fatal(err)
	}
	_, key2, err := GeneratePEM()
	if err != nil {
		t.Fatal(err)
	}
	if string(key1) == string(key2) {
		t.Error("two calls produced identical private keys")
	}
}

func TestGeneratedKeyPEMType(t *testing.T) {
	_, keyPEM, err := GeneratePEM()
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(keyPEM)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		t.Fatalf("key PEM block = %v, want RSA PRIVATE KEY", block)
	}
}
