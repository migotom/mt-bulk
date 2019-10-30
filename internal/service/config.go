package service

import (
	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/vulnerabilities"
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
	Version          string `toml:"-" yaml:"-"`
	SkipVersionCheck bool   `toml:"skip_version_check" yaml:"skip_version_check"`
	Workers          int    `toml:"workers" yaml:"workers"`
	KVStore          string `toml:"mtbulk_database" yaml:"mtbulk_database"`

	CVEURLs vulnerabilities.CVEURLs `toml:"cve_urls" yaml:"cve_urls"`
	Clients clients.Clients         `toml:"clients" yaml:"clients"`
}
