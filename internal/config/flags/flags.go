package flags

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

type Flags struct {
	ServerAddress *ServerAddress
}

type ServerAddress struct {
	Host string
	Port string
}

func (a *ServerAddress) String() string {
	value := fmt.Sprintf("%s:%s", a.Host, a.Port)
	return value
}

var errorMessage = "invalid server address. Expected format: host:port"

func (a *ServerAddress) Set(flagValue string) error {
	value := strings.Split(flagValue, ":")
	if len(value) != 2 {
		return errors.New(errorMessage)
	}
	a.Host = value[0]
	a.Port = value[1]
	if a.Host == "" || a.Port == "" {
		return errors.New(errorMessage)
	}
	return nil
}

func InitFlags() *Flags {
	flags := new(Flags)
	serverAddress := &ServerAddress{
		Host: "localhost",
		Port: "8080",
	}
	flags.ServerAddress = serverAddress
	flag.Var(serverAddress, "a", "Input server address")
	flag.Parse()

	return flags
}
