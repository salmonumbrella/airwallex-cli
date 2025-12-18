package config

import (
	"os"
	"path/filepath"
)

const AppName = "airwallex-cli"

func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, AppName), nil
}
