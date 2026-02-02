package flags

import (
	"errors"
	"flag"
	"strings"
)

type serverAddress string

type Flags struct {
	ServerAddress serverAddress
}

var errorMessage = "invalid server address. Expected format: host:port"

func InitFlags() *Flags {
	flags := new(Flags)
	var serverAddressValue serverAddress = "localhost:8080"
	flags.ServerAddress = serverAddressValue

	flag.Func("a", "Input server address", func(flagValue string) error {
		value := strings.Split(flagValue, ":")
		if len(value) != 2 {
			return errors.New(errorMessage)
		}
		host := value[0]
		port := value[1]
		if host == "" || port == "" {
			return errors.New(errorMessage)
		}
		flags.ServerAddress = serverAddress(flagValue)
		return nil
	})

	flag.Parse()

	return flags
}
