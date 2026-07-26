package settings

import (
	"flag"
	"os"
	"strconv"
)

func initEnableHTTPS(settings *Settings, cfg *fileConfig) {
	if cfg != nil && cfg.EnableHTTPS {
		settings.EnableHTTPS = true
	}

	flag.BoolFunc("s", "Enable HTTPS", func(string) error {
		settings.EnableHTTPS = true
		return nil
	})

	if env, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		if enabled, err := strconv.ParseBool(env); err == nil {
			settings.EnableHTTPS = enabled
		}
	}
}
