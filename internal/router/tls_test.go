package router

import (
	"os"
	"path/filepath"
	"testing"

	"Ustasjs/yp-url-shortener/internal/tlscert"
)

func TestGenerateTLSConfigInMemory(t *testing.T) {
	cfg, err := generateTLSConfig()
	if err != nil {
		t.Fatalf("generateTLSConfig: %v", err)
	}
	if len(cfg.Certificates) != 1 {
		t.Fatalf("got %d certificates, want 1", len(cfg.Certificates))
	}
}

func TestTLSConfigFromFiles(t *testing.T) {
	certPEM, keyPEM, err := tlscert.GeneratePEM()
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")
	keyPath := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := tlsConfigFromFiles(certPath, keyPath)
	if err != nil {
		t.Fatalf("tlsConfigFromFiles: %v", err)
	}
	if len(cfg.Certificates) != 1 {
		t.Fatalf("got %d certificates, want 1", len(cfg.Certificates))
	}
}

func TestTLSConfigFromFilesMissing(t *testing.T) {
	dir := t.TempDir()
	_, err := tlsConfigFromFiles(filepath.Join(dir, "nope-cert.pem"), filepath.Join(dir, "nope-key.pem"))
	if err == nil {
		t.Fatal("expected error for missing files")
	}
}
