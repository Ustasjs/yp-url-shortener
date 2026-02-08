package flags

import (
	"errors"
	"flag"
	"net/url"
	"strings"
)

type ServerAddress string
type BaseURL string

type Flags struct {
	ServerAddress ServerAddress
	BaseURL       BaseURL
}

func InitFlags() *Flags {
	flags := new(Flags)

	initServerAddress(flags)
	initBaseURL(flags)

	flag.Parse()

	return flags
}

var errorMessageServerAddress = "invalid server address. Expected format: host:port"

func initServerAddress(flags *Flags) {
	var serverAddressValue ServerAddress = "localhost:8080"
	flags.ServerAddress = serverAddressValue

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
		flags.ServerAddress = ServerAddress(flagValue)
		return nil
	})
}

var errorMessageBaseURL = "invalid base url. Expected format: http://host:port"

func initBaseURL(flags *Flags) {
	var baseURLvalue BaseURL = "http://localhost:8080"
	flags.BaseURL = baseURLvalue

	flag.Func("b", "Input base url", func(flagValue string) error {
		parsedURL, err := url.ParseRequestURI(flagValue)

		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			return errors.New(errorMessageBaseURL)
		}

		flags.BaseURL = BaseURL(flagValue)
		return nil
	})
}
