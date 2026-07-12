package settings

import (
	"flag"
	"os"
)

func initDatabaseDSN(settings *Settings, cfg *fileConfig) {
	if cfg != nil && cfg.DatabaseDSN != "" {
		settings.DatabaseDSN = DatabaseDSN(cfg.DatabaseDSN)
	}

	flag.Func("d", "Input database dsn", func(flagValue string) error {
		settings.DatabaseDSN = DatabaseDSN(flagValue)
		return nil
	})

	if envBaseDatabaseDsn, ok := os.LookupEnv("DATABASE_DSN"); ok {
		settings.DatabaseDSN = DatabaseDSN(envBaseDatabaseDsn)
	}
}
