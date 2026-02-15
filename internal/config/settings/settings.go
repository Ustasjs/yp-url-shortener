package settings

import (
	"errors"
	"flag"
	"net/url"
	"os"
	"strings"

	"go.uber.org/zap"
)

type ServerAddress string
type BaseURL string

type Settings struct {
	ServerAddress ServerAddress
	BaseURL       BaseURL
	LogLevel      zap.AtomicLevel
}

func InitSettings() *Settings {
	settings := new(Settings)

	initServerAddress(settings)
	initBaseURL(settings)
	initLogLevel(settings)

	flag.Parse()

	return settings
}

var errorMessageServerAddress = "invalid server address. Expected format: host:port"

func validateServerAddress(serverAddress string) error {
	value := strings.Split(serverAddress, ":")
	if len(value) != 2 {
		return errors.New(errorMessageServerAddress)
	}
	host := value[0]
	port := value[1]
	if host == "" || port == "" {
		return errors.New(errorMessageServerAddress)
	}
	return nil
}

func initServerAddress(settings *Settings) {
	var serverAddressValue ServerAddress = "localhost:8080"
	settings.ServerAddress = serverAddressValue

	flag.Func("a", "Input server address", func(flagValue string) error {
		err := validateServerAddress(flagValue)
		if err != nil {
			return err
		}
		settings.ServerAddress = ServerAddress(flagValue)
		return nil
	})

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		err := validateServerAddress(envServerAddress)
		if err != nil {
			panic(err)
		}
		settings.ServerAddress = ServerAddress(envServerAddress)
	}
}

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

func initLogLevel(settings *Settings) {
	var logLevelValue = zap.NewAtomicLevel()
	settings.LogLevel = logLevelValue

	flag.Func("l", "Input log level", func(flagValue string) error {
		lvl, err := zap.ParseAtomicLevel(flagValue)
		if err != nil {
			panic(err)
		}

		settings.LogLevel = lvl
		return nil
	})

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		lvl, err := zap.ParseAtomicLevel(envLogLevel)
		if err != nil {
			panic(err)
		}

		settings.LogLevel = lvl
	}
}
