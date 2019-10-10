package entities

import (
	"context"
	"fmt"
)

// JobsLoaderFunc is jobs loading function type.
type JobsLoaderFunc func(context.Context, Job) ([]Job, error)

// Job represents single job request.
type Job struct {
	Host     Host
	Kind     string
	Commands []Command
	Data     map[string]string
	Result   chan Result
}

func (j Job) String() string {
	return fmt.Sprintf("%s %s", j.Host, j.Kind)
}
