package settings

import (
	"errors"
	"flag"
	"net/url"
	"os"
)

var errorMessageBaseURL = "invalid base url. Expected format: http://host:port"

func validateBaseURL(baseURL string) error {
	parsedURL, err := url.ParseRequestURI(baseURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return errors.New(errorMessageBaseURL)
	}
	return nil
}

func initBaseURL(settings *Settings, cfg *fileConfig) error {
	var baseURLvalue BaseURL = "http://localhost:8080"
	settings.BaseURL = baseURLvalue

	if cfg != nil && cfg.BaseURL != "" {
		if err := validateBaseURL(cfg.BaseURL); err != nil {
			return err
		}
		settings.BaseURL = BaseURL(cfg.BaseURL)
	}

	flag.Func("b", "Input base url", func(flagValue string) error {
		err := validateBaseURL(flagValue)
		if err != nil {
			return err
		}

		settings.BaseURL = BaseURL(flagValue)
		return nil
	})

	if envBaseURL, ok := os.LookupEnv("BASE_URL"); ok {
		if err := validateBaseURL(envBaseURL); err != nil {
			return err
		}
		settings.BaseURL = BaseURL(envBaseURL)
	}
	return nil
}
