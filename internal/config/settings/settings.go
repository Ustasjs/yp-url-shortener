package settings

import (
	"errors"
	"flag"
	"net/url"
	"strings"
)

type ServerAddress string
type BaseURL string

type Settings struct {
	ServerAddress ServerAddress
	BaseURL       BaseURL
}

func InitSettings() *Settings {
	settings := new(Settings)

	initServerAddress(settings)
	initBaseURL(settings)

	flag.Parse()

	return settings
}

var errorMessageServerAddress = "invalid server address. Expected format: host:port"

func initServerAddress(settings *Settings) {
	var serverAddressValue ServerAddress = "localhost:8080"
	settings.ServerAddress = serverAddressValue

	flag.Func("a", "Input server address", func(flagValue string) error {
		value := strings.Split(flagValue, ":")
		if len(value) != 2 {
			return errors.New(errorMessageServerAddress)
		}
		host := value[0]
		port := value[1]
		if host == "" || port == "" {
			return errors.New(errorMessageServerAddress)
		}
		settings.ServerAddress = ServerAddress(flagValue)
		return nil
	})
}

var errorMessageBaseURL = "invalid base url. Expected format: http://host:port"

func initBaseURL(settings *Settings) {
	var baseURLvalue BaseURL = "http://localhost:8080"
	settings.BaseURL = baseURLvalue

	flag.Func("b", "Input base url", func(flagValue string) error {
		parsedURL, err := url.ParseRequestURI(flagValue)

		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			return errors.New(errorMessageBaseURL)
		}

		settings.BaseURL = BaseURL(flagValue)
		return nil
	})
}
