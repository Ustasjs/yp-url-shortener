package settings

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var errorMessageAuditFile = "invalid audit file path"

func validateAuditFile(filePath string) error {
	if strings.TrimSpace(filePath) == "" {
		return errors.New(errorMessageAuditFile + ": path must not be empty")
	}

	dir := filepath.Dir(filePath)

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("%s: parent directory %q does not exist: %w", errorMessageAuditFile, dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s: %q is not a directory", errorMessageAuditFile, dir)
	}

	if fileInfo, err := os.Stat(filePath); err == nil && fileInfo.IsDir() {
		return fmt.Errorf("%s: %q is a directory, not a file", errorMessageAuditFile, filePath)
	}

	return nil
}

func initAuditFile(settings *Settings) error {
	flag.Func("audit-file", "Path to audit log file (empty disables file sink)", func(flagValue string) error {
		if err := validateAuditFile(flagValue); err != nil {
			return err
		}
		settings.AuditFile = AuditFile(flagValue)
		return nil
	})

	if env, ok := os.LookupEnv("AUDIT_FILE"); ok {
		if env != "" {
			if err := validateAuditFile(env); err != nil {
				return err
			}
		}
		settings.AuditFile = AuditFile(env)
	}
	return nil
}
