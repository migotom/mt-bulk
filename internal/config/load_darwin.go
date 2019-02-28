// +build darwin

package config

import (
	"os"
	"path/filepath"
)

const configName = "config.cfg"
const configHome = ".mt-bulk.cfg"
const appSupport = "/Library/Application Support/MT-bulk"

func getConfigFileName() string {
	cfgPath := filepath.Join(os.Getenv("HOME"), appSupport, configName)
	if _, err := os.Stat(cfgPath); err == nil {
		return cfgPath
	}

	cfgPath = filepath.Join(os.Getenv("HOME"), configHome)
	if _, err := os.Stat(cfgPath); err == nil {
		return cfgPath
	}

	const appSupport = "/Library/Application Support"
	cfgPath = filepath.Join(appSupport, configName)
	if _, err := os.Stat(cfgPath); err == nil {
		return cfgPath
	}

	return ""
}
