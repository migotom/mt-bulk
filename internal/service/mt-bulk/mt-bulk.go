package mtbulk

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/migotom/mt-bulk/internal/entities"
	"github.com/migotom/mt-bulk/internal/mode"
	"github.com/migotom/mt-bulk/internal/service"
)

// MTbulk service.
type MTbulk struct {
	sugar       *zap.SugaredLogger
	jobTemplate entities.Job
	jobsLoaders []entities.JobsLoaderFunc
	jobDone     chan struct{}

	Results chan entities.Result
	Status  ApplicationStatus
	Service *service.Service
	Config
}

// NewMTbulk returns new MTbulk service.
func NewMTbulk(sugar *zap.SugaredLogger, arguments map[string]interface{}, version string) (*MTbulk, error) {
	config, jobsLoaders, jobTemplate, err := configParser(arguments, version)
	if err != nil {
		return &MTbulk{}, err
	}

	return &MTbulk{
		Config:      config,
		sugar:       sugar,
		jobsLoaders: jobsLoaders,
		jobTemplate: jobTemplate,
		jobDone:     make(chan struct{}),
		Service:     service.NewService(sugar, config.Service),
		Results:     make(chan entities.Result),
	}, nil
}

// LoadJobs loads jobs to service workers.
func (mtbulk *MTbulk) LoadJobs(ctx context.Context) {
	defer close(mtbulk.Service.Jobs)

	jobsToProcess := make([]entities.Job, 0, 256)
	if !mtbulk.Config.Service.SkipVersionCheck {
		jobsToProcess = append(jobsToProcess, entities.Job{
			Kind: mode.CheckMTbulkVersionMode,
		})
	}
	for _, jobsLoader := range mtbulk.jobsLoaders {
		jobs, err := jobsLoader(ctx, mtbulk.jobTemplate)
		if err != nil {
			mtbulk.Results <- entities.Result{Error: err}
			break
		}
		jobsToProcess = append(jobsToProcess, jobs...)
	}

	for _, job := range jobsToProcess {
		job.Result = make(chan entities.Result)

		go func(results chan entities.Result) {
			select {
			case <-ctx.Done():
			case result := <-results:
				mtbulk.Results <- result
				mtbulk.jobDone <- struct{}{}
			}
		}(job.Result)

		select {
		case <-ctx.Done():
			break
		case mtbulk.Service.Jobs <- job:
		}
	}

	go func(ctx context.Context, jobsToProcess int) {
		defer close(mtbulk.Results)
		defer close(mtbulk.jobDone)

		if jobsToProcess == 0 {
			return
		}

		jobsProcessed := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-mtbulk.jobDone:
				jobsProcessed++
				if jobsProcessed == jobsToProcess {
					return
				}
			}
		}
	}(ctx, len(jobsToProcess))
}

// ResponseCollector collects and prints out results of processed jobs.
func (mtbulk *MTbulk) ResponseCollector(ctx context.Context) {
	hostsErrors := make(map[entities.Host][]error)

collectorLooop:
	for {
		select {
		case <-ctx.Done():
			break collectorLooop
		case result, ok := <-mtbulk.Results:
			if !ok {
				break collectorLooop
			}
			if result.Error != nil {
				hostsErrors[result.Job.Host] = append(hostsErrors[result.Job.Host], result.Error)
			}

			if (mtbulk.Verbose && result.Job.Host != entities.Host{}) {
				fmt.Printf("%s > /// job: \"%s\"\n", result.Job.Host, result.Job.Kind)
				for _, commandResult := range result.Results {
					for _, response := range commandResult.Responses {
						for _, line := range strings.Split(response, "\n") {
							if line != "" {
								fmt.Printf("%s > %s\n", result.Job.Host, line)
							}
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
