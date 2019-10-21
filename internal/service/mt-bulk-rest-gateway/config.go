package mtbulkrestapi

import (
	"github.com/migotom/mt-bulk/internal/service"
)

// Config of MTbulkRESTGateway command.
type Config struct {
	Version       int            `toml:"version"`
	Listen        string         `toml:"listen" yaml:"listen"`
	RootDirectory string         `toml:"root_directory" yaml:"root_directory"`
	KeyStore      string         `toml:"keys_store" yaml:"keys_store"`
	TokenSecret   string         `toml:"token_secret" yaml:"token_secret"`
	Authenticate  []Authenticate `toml:"authenticate" yaml:"authenticate"`
	Service       service.Config `toml:"service" yaml:"service"`
}
