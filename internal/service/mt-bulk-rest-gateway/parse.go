package mtbulkrestapi

import (
	"errors"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/config"
	"github.com/migotom/mt-bulk/internal/service"
)

func configParser(arguments map[string]interface{}, version string) (mtbulkConfig Config, err error) {

	mtbulkConfig = Config{}
	mtbulkConfig.Service = service.NewConfig(version)
	mtbulkConfig.Service.Clients.SSH = clients.NewConfig(clients.SSHDefaultPort)
	mtbulkConfig.Service.Clients.MikrotikAPI = clients.NewConfig(clients.MikrotikAPIDefaultPort)

	configFileName, _ := arguments["-C"].(string)
	if err := config.LoadConfigFile(&mtbulkConfig, configFileName); err != nil {
		return Config{}, err
	}

	if mtbulkConfig.Version < 2 {
		return Config{}, errors.New("incompatible configuration version, required version 2 or above")
	}
	if mtbulkConfig.KeyStore == "" {
		return Config{}, errors.New("key store for HTTPS server not defined")
	}
	if mtbulkConfig.Listen == "" {
		return Config{}, errors.New("HTTPS server listen address not defined")
	}
	if mtbulkConfig.RootDirectory == "" {
		return Config{}, errors.New("root directory not defined")
	}

	if gen, _ := arguments["gen-https-certs"].(bool); gen {
		if err := clients.GenerateCA(mtbulkConfig.KeyStore); err != nil {
			return Config{}, err
		}
		if err := clients.GenerateCerts(mtbulkConfig.KeyStore, "rest-api"); err != nil {
			return Config{}, err
		}
	}

	return mtbulkConfig, nil
}
