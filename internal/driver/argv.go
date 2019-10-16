package driver

import (
	"context"

	"github.com/migotom/mt-bulk/internal/entities"
)

// ArgvLoadJobs loads lists of jobs using standard argument list.
func ArgvLoadJobs(ctx context.Context, jobTemplate entities.Job, args []string) (jobs []entities.Job, err error) {
	for _, entry := range args {
		job := jobTemplate
		job.Host = entities.Host{IP: entry}
		if err := job.Host.Parse(); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}
