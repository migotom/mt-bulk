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
	if gen, _ := arguments["gen-https-certs"].(bool); gen {
		if err := clients.GenerateCA(mtbulkConfig.KeyStore); err != nil {
			return Config{}, err
		}
		if err := clients.GenerateCerts(mtbulkConfig.KeyStore, "rest-gateway"); err != nil {
			return Config{}, err
		}
		return Config{}, nil
	}
	if gen, _ := arguments["gen-refresh-token"].(bool); gen {
		if err := clients.GenerateKeys(mtbulkConfig.KeyStore); err != nil {
			return Config{}, err
		}
		return Config{}, nil
	}

	return mtbulkConfig, nil
}
