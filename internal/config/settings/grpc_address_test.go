package settings

import (
	"os"
	"path/filepath"
	"testing"
)

// writeConfigFile writes a JSON config file and returns its path.
func writeConfigFile(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestGRPCAddressDefault(t *testing.T) {
	setupSettings(t)

	s, err := InitSettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(s.GRPCAddress) != "localhost:3200" {
		t.Errorf("expected default gRPC address, got %q", s.GRPCAddress)
	}
}

func TestGRPCAddressFromEnv(t *testing.T) {
	setupSettings(t)
	t.Setenv("GRPC_ADDRESS", "0.0.0.0:9090")

	s, err := InitSettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(s.GRPCAddress) != "0.0.0.0:9090" {
		t.Errorf("gRPC address not propagated: got %q", s.GRPCAddress)
	}
}

func TestGRPCAddressFromFlag(t *testing.T) {
	setupSettings(t, "-g", "localhost:9091")

	s, err := InitSettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(s.GRPCAddress) != "localhost:9091" {
		t.Errorf("gRPC address not propagated: got %q", s.GRPCAddress)
	}
}

func TestGRPCAddressFromConfigFile(t *testing.T) {
	path := writeConfigFile(t, `{"grpc_address": "localhost:9092"}`)
	setupSettings(t, "-c", path)

	s, err := InitSettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(s.GRPCAddress) != "localhost:9092" {
		t.Errorf("gRPC address not propagated: got %q", s.GRPCAddress)
	}
}

func TestGRPCAddressFlagBeatsEnv(t *testing.T) {
	setupSettings(t, "-g", "localhost:9093")
	t.Setenv("GRPC_ADDRESS", "localhost:9094")

	s, err := InitSettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(s.GRPCAddress) != "localhost:9093" {
		t.Errorf("flag must win over env: got %q", s.GRPCAddress)
	}
}

func TestGRPCAddressEnvBeatsConfigFile(t *testing.T) {
	path := writeConfigFile(t, `{"grpc_address": "localhost:9095"}`)
	setupSettings(t, "-c", path)
	t.Setenv("GRPC_ADDRESS", "localhost:9096")

	s, err := InitSettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(s.GRPCAddress) != "localhost:9096" {
		t.Errorf("env must win over config file: got %q", s.GRPCAddress)
	}
}

// A host-only address listens on every interface, so it must be accepted.
func TestGRPCAddressPortOnly(t *testing.T) {
	setupSettings(t)
	t.Setenv("GRPC_ADDRESS", ":3200")

	s, err := InitSettings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(s.GRPCAddress) != ":3200" {
		t.Errorf("gRPC address not propagated: got %q", s.GRPCAddress)
	}
}

func TestGRPCAddressInvalidIsError(t *testing.T) {
	for _, value := range []string{"localhost", "localhost:", "", "a:b:c"} {
		setupSettings(t)
		t.Setenv("GRPC_ADDRESS", value)

		if _, err := InitSettings(); err == nil {
			t.Errorf("expected error for gRPC address %q", value)
		}
	}
}
