// Package settings loads the application configuration from command-line flags
// and environment variables into a single Settings value.
package settings

import (
	"flag"

	"go.uber.org/zap"
)

type ServerAddress string
type BaseURL string
type FileStoragePath string
type DatabaseDSN string
type AuditFile string
type AuditURL string
type TLSCertFile string
type TLSKeyFile string
type TrustedSubnet string

type Settings struct {
	ServerAddress   ServerAddress
	BaseURL         BaseURL
	LogLevel        zap.AtomicLevel
	FileStoragePath FileStoragePath
	DatabaseDSN     DatabaseDSN
	AuditFile       AuditFile
	AuditURL        AuditURL
	EnableHTTPS     bool
	TLSCertFile     TLSCertFile
	TLSKeyFile      TLSKeyFile
	TrustedSubnet   TrustedSubnet
}

func InitSettings() (*Settings, error) {
	settings := new(Settings)
	// Ensure a valid log level even if we fail early, so the caller can log the
	// error instead of panicking on a zero-value AtomicLevel.
	settings.LogLevel = zap.NewAtomicLevel()

	cfg, err := loadFileConfig()
	if err != nil {
		return settings, err
	}
	registerConfigFlags()

	if err := initLogLevel(settings, cfg); err != nil {
		return settings, err
	}
	if err := initServerAddress(settings, cfg); err != nil {
		return settings, err
	}
	if err := initBaseURL(settings, cfg); err != nil {
		return settings, err
	}
	if err := initFileStoragePath(settings, cfg); err != nil {
		return settings, err
	}
	initDatabaseDSN(settings, cfg)
	if err := initAuditFile(settings, cfg); err != nil {
		return settings, err
	}
	if err := initAuditURL(settings, cfg); err != nil {
		return settings, err
	}
	initEnableHTTPS(settings, cfg)
	initTLSFiles(settings, cfg)
	if err := initTrustedSubnet(settings, cfg); err != nil {
		return settings, err
	}

	flag.Parse()

	// TLS file paths need cross-field ("both or neither") validation, so it runs
	// after flag.Parse once every source has been merged.
	if err := validateTLSFiles(settings); err != nil {
		return settings, err
	}

	return settings, nil
}
