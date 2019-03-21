package config

import (
	"errors"

	"github.com/BurntSushi/toml"
)

// LoadConfigFile loads config file using per system defined tree, eg. Linux: home, etc
func LoadConfigFile(config interface{}, implicitConfigFileName string) error {
	file, err := getFileName(implicitConfigFileName)
	if err != nil {
		return err
	}

	if _, err := toml.DecodeFile(file, config); err != nil {
		return err
	}
	return nil
}

func getFileName(implicitConfigFileName string) (string, error) {
	fileName := getConfigFileName()
	if implicitConfigFileName != "" {
		fileName = implicitConfigFileName
	}

	if fileName == "" {
		return "", errors.New("no config file provided")
	}

	return fileName, nil
}
