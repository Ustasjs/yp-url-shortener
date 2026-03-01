package settings

import (
	"flag"
	"os"
)

func initDatabaseDSN(settings *Settings) {
	flag.Func("d", "Input database dsn", func(flagValue string) error {
		settings.DatabaseDSN = DatabaseDSN(flagValue)
		return nil
	})

	if envBaseDatabaseDsn := os.Getenv("DATABASE_DSN"); envBaseDatabaseDsn != "" {
		settings.DatabaseDSN = DatabaseDSN(envBaseDatabaseDsn)
	}
}
