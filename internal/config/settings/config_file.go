package settings

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// fileConfig mirrors the JSON configuration file. Every field maps to an
// existing command-line flag / environment variable. Values loaded here act as
// the lowest-priority configuration source (above hardcoded defaults only).
type fileConfig struct {
	ServerAddress   string `json:"server_address"`
	GRPCAddress     string `json:"grpc_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
	LogLevel        string `json:"log_level"`
	AuditFile       string `json:"audit_file"`
	AuditURL        string `json:"audit_url"`
	TLSCertFile     string `json:"tls_cert_file"`
	TLSKeyFile      string `json:"tls_key_file"`
	TrustedSubnet   string `json:"trusted_subnet"`
}

// registerConfigFlags declares -c/-config so flag.Parse does not fail on them
// and they show up in the -h usage. The actual path is resolved earlier by
// resolveConfigPath, so these handlers only need to accept the value.
func registerConfigFlags() {
	flag.Func("c", "Path to JSON config file", func(string) error { return nil })
	flag.Func("config", "Path to JSON config file", func(string) error { return nil })
}

// resolveConfigPath determines the config file path before flag.Parse runs, so
// its values can be used as baselines. The -c/-config flag takes precedence
// over the CONFIG environment variable.
func resolveConfigPath() string {
	path := ""
	if env, ok := os.LookupEnv("CONFIG"); ok {
		path = env
	}

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		for _, name := range []string{"-c", "--c", "-config", "--config"} {
			switch {
			case arg == name:
				if i+1 < len(args) {
					path = args[i+1]
				}
			case strings.HasPrefix(arg, name+"="):
				path = strings.TrimPrefix(arg, name+"=")
			}
		}
	}

	return path
}

// loadFileConfig resolves and reads the JSON config file. It returns (nil, nil)
// when no config path is provided, and an error when the file cannot be read or
// parsed so a broken config is surfaced rather than silently ignored.
func loadFileConfig() (*fileConfig, error) {
	path := resolveConfigPath()
	if path == "" {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file %q: %w", path, err)
	}

	cfg := new(fileConfig)
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file %q: %w", path, err)
	}

	return cfg, nil
}
