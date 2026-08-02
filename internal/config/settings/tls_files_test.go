package settings

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// relevantEnv lists every environment variable the settings package reads, so a
// test can start from a clean, predictable environment.
var relevantEnv = []string{
	"SERVER_ADDRESS", "GRPC_ADDRESS", "BASE_URL", "DATABASE_DSN",
	"FILE_STORAGE_PATH", "LOG_LEVEL", "ENABLE_HTTPS", "AUDIT_FILE",
	"AUDIT_URL", "TLS_CERT_FILE", "TLS_KEY_FILE", "CONFIG", "TRUSTED_SUBNET",
}

// setupSettings gives each case a fresh flag set and argv (InitSettings
// registers flags on the global flag.CommandLine) and a cleared environment.
func setupSettings(t *testing.T, args ...string) {
	t.Helper()
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	os.Args = append([]string{"test"}, args...)
	for _, env := range relevantEnv {
		os.Unsetenv(env)
	}
}

func writeTempFile(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestTLSFilesNeitherSet(t *testing.T) {
	setupSettings(t)

	s, err := InitSettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.TLSCertFile != "" || s.TLSKeyFile != "" {
		t.Errorf("expected empty TLS paths, got cert=%q key=%q", s.TLSCertFile, s.TLSKeyFile)
	}
}

func TestTLSFilesOnlyCertIsError(t *testing.T) {
	cert := writeTempFile(t, "cert.pem")
	setupSettings(t)
	t.Setenv("TLS_CERT_FILE", cert)

	if _, err := InitSettings(); err == nil {
		t.Fatal("expected error when only cert-file is set")
	}
}

func TestTLSFilesOnlyKeyIsError(t *testing.T) {
	key := writeTempFile(t, "key.pem")
	setupSettings(t)
	t.Setenv("TLS_KEY_FILE", key)

	if _, err := InitSettings(); err == nil {
		t.Fatal("expected error when only key-file is set")
	}
}

func TestTLSFilesMissingFileIsError(t *testing.T) {
	setupSettings(t)
	t.Setenv("TLS_CERT_FILE", filepath.Join(t.TempDir(), "nope-cert.pem"))
	t.Setenv("TLS_KEY_FILE", filepath.Join(t.TempDir(), "nope-key.pem"))

	if _, err := InitSettings(); err == nil {
		t.Fatal("expected error when TLS files do not exist")
	}
}

func TestTLSFilesBothPresentFromEnv(t *testing.T) {
	cert := writeTempFile(t, "cert.pem")
	key := writeTempFile(t, "key.pem")
	setupSettings(t)
	t.Setenv("TLS_CERT_FILE", cert)
	t.Setenv("TLS_KEY_FILE", key)

	s, err := InitSettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(s.TLSCertFile) != cert || string(s.TLSKeyFile) != key {
		t.Errorf("paths not propagated: cert=%q key=%q", s.TLSCertFile, s.TLSKeyFile)
	}
}

func TestTLSFilesFromFlags(t *testing.T) {
	cert := writeTempFile(t, "cert.pem")
	key := writeTempFile(t, "key.pem")
	setupSettings(t, "-cert-file", cert, "-key-file", key)

	s, err := InitSettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(s.TLSCertFile) != cert || string(s.TLSKeyFile) != key {
		t.Errorf("flag paths not propagated: cert=%q key=%q", s.TLSCertFile, s.TLSKeyFile)
	}
}

func TestEnableHTTPSEnvParsing(t *testing.T) {
	cases := map[string]bool{"true": true, "false": false, "1": true, "0": false}
	for value, want := range cases {
		setupSettings(t)
		t.Setenv("ENABLE_HTTPS", value)

		s, err := InitSettings()
		if err != nil {
			t.Fatalf("ENABLE_HTTPS=%q: unexpected error: %v", value, err)
		}
		if s.EnableHTTPS != want {
			t.Errorf("ENABLE_HTTPS=%q -> %v, want %v", value, s.EnableHTTPS, want)
		}
	}
}
