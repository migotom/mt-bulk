package mtbulk

import (
	"github.com/migotom/mt-bulk/internal/driver"
	"github.com/migotom/mt-bulk/internal/entities"
	"github.com/migotom/mt-bulk/internal/service"
)

// Config of MTbulk command.
type Config struct {
	Version     int  `toml:"version"`
	Verbose     bool `toml:"verbose"`
	SkipSummary bool `toml:"skip_summary" yaml:"skip_summary"`

	Service           service.Config  `toml:"service" yaml:"service"`
	DB                driver.DBConfig `toml:"db" yaml:"db"`
	CustomSSHSequence *CustomSequence `toml:"custom-ssh" yaml:"custom-ssh"`
	CustomAPISequence *CustomSequence `toml:"custom-api" yaml:"custom-api"`
}

// CustomSequence is sequence of custom commands.
type CustomSequence struct {
	Command []entities.Command `toml:"command" yaml:"command"`
}
