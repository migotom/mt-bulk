package service

import (
	"github.com/migotom/mt-bulk/internal/clients"
)

// NewConfig returns new service config.
func NewConfig(version string) Config {
	return Config{
		Version: version,
		Workers: 4,
	}
}

// Config is service configuration.
type Config struct {
	Version          string `toml:"-"`
	SkipVersionCheck bool   `toml:"skip_version_check"`
	Workers          int    `toml:"workers"`

	Clients clients.Clients `toml:"clients"`
}
