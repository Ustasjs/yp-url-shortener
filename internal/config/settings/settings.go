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

type Settings struct {
	ServerAddress   ServerAddress
	BaseURL         BaseURL
	LogLevel        zap.AtomicLevel
	FileStoragePath FileStoragePath
	DatabaseDSN     DatabaseDSN
	AuditFile       AuditFile
	AuditURL        AuditURL
}

func InitSettings() (*Settings, error) {
	settings := new(Settings)

	if err := initLogLevel(settings); err != nil {
		return settings, err
	}
	if err := initServerAddress(settings); err != nil {
		return settings, err
	}
	if err := initBaseURL(settings); err != nil {
		return settings, err
	}
	if err := initFileStoragePath(settings); err != nil {
		return settings, err
	}
	initDatabaseDSN(settings)
	if err := initAuditFile(settings); err != nil {
		return settings, err
	}
	if err := initAuditURL(settings); err != nil {
		return settings, err
	}

	flag.Parse()

	return settings, nil
}
