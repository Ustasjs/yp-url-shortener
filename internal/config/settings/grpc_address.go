package settings

import (
	"errors"
	"flag"
	"net"
	"os"
)

var errorMessageGRPCAddress = "invalid gRPC address. Expected format: host:port"

// validateGRPCAddress uses net.SplitHostPort, so both "host:port" and ":port"
// pass. An empty host means "listen on every interface", which is normal for
// gRPC.
func validateGRPCAddress(grpcAddress string) error {
	_, port, err := net.SplitHostPort(grpcAddress)
	if err != nil || port == "" {
		return errors.New(errorMessageGRPCAddress)
	}
	return nil
}

func initGRPCAddress(settings *Settings, cfg *fileConfig) error {
	var grpcAddressValue GRPCAddress = "localhost:3200"
	settings.GRPCAddress = grpcAddressValue

	if cfg != nil && cfg.GRPCAddress != "" {
		if err := validateGRPCAddress(cfg.GRPCAddress); err != nil {
			return err
		}
		settings.GRPCAddress = GRPCAddress(cfg.GRPCAddress)
	}

	flag.Func("g", "Input gRPC server address", func(flagValue string) error {
		err := validateGRPCAddress(flagValue)
		if err != nil {
			return err
		}
		settings.GRPCAddress = GRPCAddress(flagValue)
		return nil
	})

	if envGRPCAddress, ok := os.LookupEnv("GRPC_ADDRESS"); ok {
		if err := validateGRPCAddress(envGRPCAddress); err != nil {
			return err
		}
		settings.GRPCAddress = GRPCAddress(envGRPCAddress)
	}
	return nil
}
