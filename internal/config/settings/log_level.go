package settings

import (
	"flag"
	"os"

	"go.uber.org/zap"
)

func initLogLevel(settings *Settings) error {
	var logLevelValue = zap.NewAtomicLevel()
	settings.LogLevel = logLevelValue

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
