package settings

import (
	"flag"
	"os"

	"go.uber.org/zap"
)

func initLogLevel(settings *Settings, cfg *fileConfig) error {
	var logLevelValue = zap.NewAtomicLevel()
	settings.LogLevel = logLevelValue

	if cfg != nil && cfg.LogLevel != "" {
		lvl, err := zap.ParseAtomicLevel(cfg.LogLevel)
		if err != nil {
			return err
		}
		settings.LogLevel = lvl
	}

	flag.Func("l", "Input log level", func(flagValue string) error {
		lvl, err := zap.ParseAtomicLevel(flagValue)
		if err != nil {
			return err
		}

		settings.LogLevel = lvl
		return nil
	})

	if envLogLevel, ok := os.LookupEnv("LOG_LEVEL"); ok {
		lvl, err := zap.ParseAtomicLevel(envLogLevel)
		if err != nil {
			return err
		}

		settings.LogLevel = lvl
	}
	return nil
}
