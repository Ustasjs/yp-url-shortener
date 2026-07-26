package settings

import (
	"flag"
	"fmt"
	"net"
	"os"
)

func validateTrustedSubnet(raw string) error {
	if _, _, err := net.ParseCIDR(raw); err != nil {
		return fmt.Errorf("invalid trusted subnet: %q is not a valid CIDR", raw)
	}
	return nil
}

// initTrustedSubnet resolves the trusted subnet (CIDR) from the config file,
// the -t flag and the TRUSTED_SUBNET environment variable, in ascending order
// of priority. An empty value is allowed and disables access to the internal
// stats endpoint for every client; a non-empty value must be valid CIDR.
func initTrustedSubnet(settings *Settings, cfg *fileConfig) error {
	if cfg != nil && cfg.TrustedSubnet != "" {
		if err := validateTrustedSubnet(cfg.TrustedSubnet); err != nil {
			return err
		}
		settings.TrustedSubnet = TrustedSubnet(cfg.TrustedSubnet)
	}

	flag.Func("t", "Trusted subnet in CIDR notation (empty forbids /api/internal/stats)", func(flagValue string) error {
		if flagValue != "" {
			if err := validateTrustedSubnet(flagValue); err != nil {
				return err
			}
		}
		settings.TrustedSubnet = TrustedSubnet(flagValue)
		return nil
	})

	if env, ok := os.LookupEnv("TRUSTED_SUBNET"); ok {
		if env != "" {
			if err := validateTrustedSubnet(env); err != nil {
				return err
			}
		}
		settings.TrustedSubnet = TrustedSubnet(env)
	}
	return nil
}
