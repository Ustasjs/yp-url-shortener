package settings

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
)

var errorMessageAuditURL = "invalid audit url"

func validateAuditURL(raw string) error {
	if strings.TrimSpace(raw) == "" {
		return errors.New(errorMessageAuditURL + ": url must not be empty")
	}
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("%s: %q is not a valid absolute URL", errorMessageAuditURL, raw)
	}
	return nil
}

func initAuditURL(settings *Settings) error {
	flag.Func("audit-url", "Remote audit sink URL (empty disables HTTP sink)", func(flagValue string) error {
		if err := validateAuditURL(flagValue); err != nil {
			return err
		}
		settings.AuditURL = AuditURL(flagValue)
		return nil
	})

	if env := os.Getenv("AUDIT_URL"); env != "" {
		if err := validateAuditURL(env); err != nil {
			return err
		}
		settings.AuditURL = AuditURL(env)
	}
	return nil
}
