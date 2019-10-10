package driver

import (
	"bufio"
	"context"
	"os"

	"github.com/migotom/mt-bulk/internal/entities"
)

// FileLoadJobs loads list of jobs from file.
func FileLoadJobs(ctx context.Context, jobTemplate entities.Job, filename string) (jobs []entities.Job, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		job := jobTemplate
		job.Host = entities.Host{IP: scanner.Text()}
		if err := job.Host.Parse(); err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}
