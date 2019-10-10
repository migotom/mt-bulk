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
	SkipSummary bool `toml:"skip_summary"`

	Service           service.Config
	DB                driver.DBConfig
	CustomSSHSequence *CustomSequence `toml:"custom-ssh"`
	CustomAPISequence *CustomSequence `toml:"custom-api"`
}

// CustomSequence is sequence of custom commands.
type CustomSequence struct {
	Command []entities.Command
}
