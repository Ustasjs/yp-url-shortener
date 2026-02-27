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

func initBaseURL(settings *Settings) {
	var baseURLvalue BaseURL = "http://localhost:8080"
	settings.BaseURL = baseURLvalue

	flag.Func("b", "Input base url", func(flagValue string) error {
		err := validateBaseURL(flagValue)
		if err != nil {
			return err
		}

		settings.BaseURL = BaseURL(flagValue)
		return nil
	})

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		err := validateBaseURL(envBaseURL)
		if err != nil {
			panic(err)
		}
		settings.BaseURL = BaseURL(envBaseURL)
	}
}
