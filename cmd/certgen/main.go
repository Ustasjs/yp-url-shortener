// Command certgen generates a self-signed TLS certificate and private key and
// writes them to PEM files. It reuses the certificate-generation logic shared
// with the HTTPS server (internal/tlscert), so files produced here are
// equivalent to the certificate the server generates in memory.
//
// Usage:
//
//	go run ./cmd/certgen [-cert cert.pem] [-key key.pem]
//
// The service can then be started with the generated files, e.g.:
//
//	go run ./cmd/shortener -s -cert-file cert.pem -key-file key.pem
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"Ustasjs/yp-url-shortener/internal/tlscert"
)

func main() {
	certPath := flag.String("cert", "cert.pem", "Output path for the certificate (PEM)")
	keyPath := flag.String("key", "key.pem", "Output path for the private key (PEM)")
	flag.Parse()

	certPEM, keyPEM, err := tlscert.GeneratePEM()
	if err != nil {
		log.Fatalf("generate certificate: %v", err)
	}

	// The certificate is public; the private key must stay owner-readable only.
	if err := os.WriteFile(*certPath, certPEM, 0o644); err != nil {
		log.Fatalf("write certificate %q: %v", *certPath, err)
	}
	if err := os.WriteFile(*keyPath, keyPEM, 0o600); err != nil {
		log.Fatalf("write private key %q: %v", *keyPath, err)
	}

	fmt.Printf("Wrote certificate to %s\n", *certPath)
	fmt.Printf("Wrote private key to %s\n", *keyPath)
}
