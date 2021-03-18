package mtbulk

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/migotom/mt-bulk/internal/entities"
	"github.com/migotom/mt-bulk/internal/kvdb"
	"github.com/migotom/mt-bulk/internal/mode"
	"github.com/migotom/mt-bulk/internal/service"
	"github.com/migotom/mt-bulk/internal/vulnerabilities"

	"github.com/rs/xid"
	"go.uber.org/zap"
)

// MTbulk service.
type MTbulk struct {
	sugar       *zap.SugaredLogger
	kv          kvdb.KV
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
		return &MTbulk{}, fmt.Errorf("configuration parser:%s", err)
	}

	if reflect.DeepEqual(entities.Job{}, jobTemplate) {
		return nil, nil
	}

	kv, err := kvdb.OpenKV(sugar, config.Service.KVStore)
	if err != nil {
		return &MTbulk{}, fmt.Errorf("creating cache KV store:%s", err)
	}

	return &MTbulk{
		Config:      config,
		sugar:       sugar,
		kv:          kv,
		jobsLoaders: jobsLoaders,
		jobTemplate: jobTemplate,
		jobDone:     make(chan struct{}),
		Service:     service.NewService(sugar, kv, config.Service),
		Results:     make(chan entities.Result),
	}, nil
}

// LoadJobs loads jobs to service workers.
func (mtbulk *MTbulk) LoadJobs(ctx context.Context) {
	defer close(mtbulk.Service.Jobs)
	defer close(mtbulk.Results)

	jobsToProcess := make([]entities.Job, 0, 256)
	if !mtbulk.Config.Service.SkipVersionCheck {
		jobsToProcess = append(jobsToProcess, entities.Job{
			Kind: mode.CheckMTbulkVersionMode,
		})
	}
	for _, jobsLoader := range mtbulk.jobsLoaders {
		jobs, err := jobsLoader(ctx, mtbulk.jobTemplate)
		if err != nil {
			mtbulk.Results <- entities.Result{Errors: []error{err}}
			break
		}
		jobsToProcess = append(jobsToProcess, jobs...)
	}

	wgJobs := new(sync.WaitGroup)
	for _, job := range jobsToProcess {
		wgJobs.Add(1)
		go func(job entities.Job) {
			defer wgJobs.Done()

			job.ID = xid.New().String()
			job.Result = make(chan entities.Result)

			select {
			case <-ctx.Done():
			case mtbulk.Service.Jobs <- job:
			}

			select {
			case <-ctx.Done():
			case result := <-job.Result:
				mtbulk.Results <- result
			}
		}(job)
	}
	wgJobs.Wait()
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
			if result.Errors != nil {
				hostsErrors[result.Job.Host] = append(hostsErrors[result.Job.Host], result.Errors...)
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

	vulnerabilitiesDetected := false

	fmt.Println()
	fmt.Println("Errors list:")
	for host, errors := range hostsErrors {
		if host.IP != "" {
			fmt.Printf("Device: %s:%s\n", host.IP, host.Port)
		} else {
			fmt.Println("Generic:")
		}

		for _, err := range errors {
			if err == nil {
				continue
			}
			fmt.Printf("\t%s\n", err)
			if vul, ok := err.(vulnerabilities.VulnerabilityError); ok && vul.Vulnerabilities != nil {
				vulnerabilitiesDetected = true
			}
		}
	}

	if vulnerabilitiesDetected {
		cves := ExtractCVEs(hostsErrors)

		fmt.Println()
		fmt.Println("Deetected CVE:")
		for _, cve := range cves {
			fmt.Printf("%s\n", cve)
		}
	}
}

// Listen runs service workers and process all provided jobs.
// Returns after process of all jobs.
func (mtbulk *MTbulk) Listen(ctx context.Context, cancel context.CancelFunc) {
	defer mtbulk.kv.Close()

	mtbulk.Service.Listen(ctx, cancel)
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

// ExtractCVEs extracts list of unique CVEs from list of hosts' errors.
func ExtractCVEs(hostsErrors map[entities.Host][]error) []vulnerabilities.CVE {
	cves := make([]vulnerabilities.CVE, 0, len(hostsErrors))

	for _, errors := range hostsErrors {
		for _, err := range errors {
			vul, ok := err.(vulnerabilities.VulnerabilityError)
			if !ok {
				continue
			}

		vulnerabilitiesScan:
			for _, cve := range vul.Vulnerabilities {
				for _, storedCve := range cves {
					if storedCve.ID == cve.ID {
						continue vulnerabilitiesScan
					}
				}
				cves = append(cves, cve)
			}
		}

	}
	return cves
}
