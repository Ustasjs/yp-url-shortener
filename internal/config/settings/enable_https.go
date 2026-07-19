package settings

import (
	"flag"
	"os"
)

func initEnableHTTPS(settings *Settings, cfg *fileConfig) {
	if cfg != nil && cfg.EnableHTTPS {
		settings.EnableHTTPS = true
	}

	flag.BoolFunc("s", "Enable HTTPS", func(string) error {
		settings.EnableHTTPS = true
		return nil
	})

	if _, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		settings.EnableHTTPS = true
	}
}
