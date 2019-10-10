package entities

import (
	"context"
	"fmt"
)

// JobsLoaderFunc is jobs loading function type.
type JobsLoaderFunc func(context.Context, Job) ([]Job, error)

// Job represents single job request.
type Job struct {
	Host     Host              `toml:"host" yaml:"host"`
	Kind     string            `toml:"kind" yaml:"kind"`
	Commands []Command         `toml:"commands" yaml:"commands"`
	Data     map[string]string `toml:"data"  yaml:"data"`
	Result   chan Result       `toml:"result" yaml:"result"`
}

func (j Job) String() string {
	return fmt.Sprintf("%s %s", j.Host, j.Kind)
}
