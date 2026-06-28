package settings

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var errorMessageFileStoragePath = "invalid file storage path"

func validateFileStoragePath(filePath string) error {
	if strings.TrimSpace(filePath) == "" {
		return errors.New(errorMessageFileStoragePath + ": path must not be empty")
	}

	dir := filepath.Dir(filePath)

	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("%s: parent directory %q does not exist: %w", errorMessageFileStoragePath, dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s: %q is not a directory", errorMessageFileStoragePath, dir)
	}

	if fileInfo, err := os.Stat(filePath); err == nil && fileInfo.IsDir() {
		return fmt.Errorf("%s: %q is a directory, not a file", errorMessageFileStoragePath, filePath)
	}

	return nil
}

func initFileStoragePath(settings *Settings) error {
	var fileStoragePathValue FileStoragePath = "./file_storage.json"
	settings.FileStoragePath = fileStoragePathValue

	flag.Func("f", "Input file storage path", func(flagValue string) error {
		err := validateFileStoragePath(flagValue)
		if err != nil {
			return err
		}
		settings.FileStoragePath = FileStoragePath(flagValue)
		return nil
	})

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		err := validateFileStoragePath(envFileStoragePath)
		if err != nil {
			return err
		}
		settings.FileStoragePath = FileStoragePath(envFileStoragePath)
	}
	return nil
}
