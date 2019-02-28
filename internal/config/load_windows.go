// +build !linux,!darwin

package config

import (
	"os"
	"path/filepath"
)

const configHome = "mt-bulk.cfg"
const configSys = "config.cfg"

const etc = "/etc/mt-bulk/"

func getConfigFileName() string {
	cfgPath := filepath.Join(os.Getenv("HOME"), configHome)
	if _, err := os.Stat(cfgPath); err == nil {
		return cfgPath
	}

	cfgPath = filepath.Join(etc, configSys)
	if _, err := os.Stat(cfgPath); err == nil {
		return cfgPath
	}

	return ""
}
