package settings

import (
	"flag"

	"go.uber.org/zap"
)

type ServerAddress string
type BaseURL string
type FileStoragePath string

type Settings struct {
	ServerAddress   ServerAddress
	BaseURL         BaseURL
	LogLevel        zap.AtomicLevel
	FileStoragePath FileStoragePath
}

func InitSettings() *Settings {
	settings := new(Settings)

	initServerAddress(settings)
	initBaseURL(settings)
	initLogLevel(settings)
	initFileStoragePath(settings)

	flag.Parse()

	return settings
}
