package config

import (
	"github.com/BurntSushi/toml"
	"github.com/migotom/mt-bulk/internal/schema"
)

// LoadConfigFile loads config file using per system defined tree, eg. Linux: home, etc
func LoadConfigFile(config *schema.GeneralConfig, implicitConfigFileName string) error {
	fileName := getConfigFileName()
	if implicitConfigFileName != "" {
		fileName = implicitConfigFileName
	}

	if fileName != "" {
		if _, err := toml.DecodeFile(fileName, &config); err != nil {
			return err
		}
	}

	return nil
}
