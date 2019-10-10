package mtbulk

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/migotom/mt-bulk/internal/entities"
	"github.com/migotom/mt-bulk/internal/service"
)

// MTbulk service.
type MTbulk struct {
	jobTemplate entities.Job
	jobsLoaders []entities.JobsLoaderFunc

	Status  ApplicationStatus
	Service *service.Service
	Config
}

// NewMTbulk returns new MTbulk service.
func NewMTbulk(arguments map[string]interface{}, version string) (MTbulk, error) {
	config, jobsLoaders, jobTemplate, err := configParser(arguments, version)
	if err != nil {
		return MTbulk{}, err
	}

	return MTbulk{
		Config:      config,
		jobsLoaders: jobsLoaders,
		jobTemplate: jobTemplate,
		Service:     service.NewService(config.Service),
	}, nil
}

// LoadJobs loads jobs to service workers.
func (mtbulk *MTbulk) LoadJobs(ctx context.Context) {
	defer close(mtbulk.Service.Jobs)

	mtbulk.jobTemplate.Result = mtbulk.Service.Results
	jobsToProcess := make([]entities.Job, 0, 256)

	for _, jobsLoader := range mtbulk.jobsLoaders {
		jobs, err := jobsLoader(ctx, mtbulk.jobTemplate)
		if err != nil {
			mtbulk.Service.Results <- entities.Result{Error: err}
			return
		}
		jobsToProcess = append(jobsToProcess, jobs...)
	}

	for _, job := range jobsToProcess {
		select {
		case <-ctx.Done():
			mtbulk.Service.Results <- entities.Result{Error: ctx.Err()}
			return
		case mtbulk.Service.Jobs <- job:
		}
	}

}

// ResponseCollector collects and prints out results of processed jobs.
func (mtbulk *MTbulk) ResponseCollector(ctx context.Context) {
	hostsErrors := make(map[entities.Host][]error)

collectorLooop:
	for {
		select {
		case <-ctx.Done():

			break collectorLooop
		case result, ok := <-mtbulk.Service.Results:
			if !ok {
				break collectorLooop
			}
			if result.Error != nil {
				hostsErrors[result.Host] = append(hostsErrors[result.Host], result.Error)
			}

			if mtbulk.Verbose {
				fmt.Printf("%s > /// job: \"%s\"\n", result.Host, result.Job.Kind)
				for _, response := range result.Responses {
					for _, line := range strings.Split(response, "\n") {
						if line != "" {
							fmt.Printf("%s > %s\n", result.Host, line)
						}
					}
				}
			}
		}
	}
	if len(hostsErrors) == 0 {
		return
	}

	mtbulk.Status.SetCode(1)

	if mtbulk.SkipSummary {
		return
	}

	fmt.Println()
	fmt.Println("Errors list:")
	for host, errors := range hostsErrors {
		if host.IP != "" {
			fmt.Printf("Device: %s:%s\n", host.IP, host.Port)
		}

		for _, error := range errors {
			fmt.Printf("\t%s\n", error)
		}
	}
}

// Listen runs service workers and process all provided jobs.
// Returns after process of all jobs.
func (mtbulk *MTbulk) Listen(ctx context.Context) {
	mtbulk.Service.Listen(ctx)
}

// ApplicationStatus stores final execution status that should be returned to OS.
type ApplicationStatus struct {
	code int
	sync.Mutex
}

// SetCode sets application status code.
func (app *ApplicationStatus) SetCode(status int) {
	app.Lock()
	defer app.Unlock()

	app.code = status
}

// Get returns application status code.
func (app *ApplicationStatus) Get() int {
	app.Lock()
	defer app.Unlock()

	return app.code
}
