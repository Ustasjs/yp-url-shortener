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

func InitSettings() *Settings {
	settings := new(Settings)

	initServerAddress(settings)
	initBaseURL(settings)
	initLogLevel(settings)
	initFileStoragePath(settings)
	initDatabaseDSN(settings)
	initAuditFile(settings)
	initAuditURL(settings)

	flag.Parse()

	return settings
}
