package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

// LoadConfigFile loads config file using per system defined tree, eg. Linux: home, etc
func LoadConfigFile(config interface{}, implicitConfigFileName string) error {
	file, err := getFileName(implicitConfigFileName)
	if err != nil {
		return err
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	switch strings.ToLower(filepath.Ext(file)) {
	case ".toml":
		err = toml.Unmarshal(content, config)
		if err != nil {
			return err
		}
	case ".yml", ".yaml":
		err = yaml.Unmarshal(content, config)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("not supported configuration type %s, MT-bulk supports .toml, .yml and .yaml files", file)
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
