package settings

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

// initTLSFiles wires the TLS certificate/key paths from the config file, flags
// and environment into settings. The values are validated later by
// validateTLSFiles, after flag.Parse has merged every source
func initTLSFiles(settings *Settings, cfg *fileConfig) {
	if cfg != nil {
		if cfg.TLSCertFile != "" {
			settings.TLSCertFile = TLSCertFile(cfg.TLSCertFile)
		}
		if cfg.TLSKeyFile != "" {
			settings.TLSKeyFile = TLSKeyFile(cfg.TLSKeyFile)
		}
	}

	flag.Func("cert-file", "Path to TLS certificate file (PEM); requires -key-file", func(v string) error {
		settings.TLSCertFile = TLSCertFile(v)
		return nil
	})
	flag.Func("key-file", "Path to TLS private key file (PEM); requires -cert-file", func(v string) error {
		settings.TLSKeyFile = TLSKeyFile(v)
		return nil
	})

	if env, ok := os.LookupEnv("TLS_CERT_FILE"); ok {
		settings.TLSCertFile = TLSCertFile(env)
	}
	if env, ok := os.LookupEnv("TLS_KEY_FILE"); ok {
		settings.TLSKeyFile = TLSKeyFile(env)
	}
}

// validateTLSFiles enforces that the certificate and key paths are provided
// together and that each points to an existing file.
func validateTLSFiles(settings *Settings) error {
	cert := string(settings.TLSCertFile)
	key := string(settings.TLSKeyFile)

	if (cert == "") != (key == "") {
		return errors.New("incomplete TLS configuration: cert-file and key-file must be set together")
	}
	if cert == "" {
		return nil
	}

	if err := ensureRegularFile(cert, "TLS certificate"); err != nil {
		return err
	}
	return ensureRegularFile(key, "TLS key")
}

func ensureRegularFile(path, label string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s file %q: %w", label, path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s file %q is a directory, not a file", label, path)
	}
	return nil
}
