package service

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/migotom/mt-bulk/internal/clients"
	"github.com/migotom/mt-bulk/internal/entities"
	"github.com/migotom/mt-bulk/internal/mode"
)

// WorkerPool of SSH/Mikrotik SSL API clients.
type WorkerPool struct {
	sync.Mutex

	current int
	pool    []*Worker
}

// NewWorkerPool returns new worker pool.
func NewWorkerPool(numberOfWorkers int) *WorkerPool {
	return &WorkerPool{
		current: 0,
		pool:    make([]*Worker, 0, numberOfWorkers),
	}
}

// Close all workers from worker pool.
func (p *WorkerPool) Close() {
	p.Lock()
	defer p.Unlock()

	for _, worker := range p.pool {
		close(worker.jobs)
	}
}

// Add new worker to worker pool.
func (p *WorkerPool) Add(worker *Worker) {
	p.Lock()
	defer p.Unlock()

	p.pool = append(p.pool, worker)
}

// Get worker that is already processing jobs for given host or if not found any pick one using round robin.
func (p *WorkerPool) Get(host entities.Host) (worker *Worker) {
	p.Lock()
	defer p.Unlock()
	for _, worker := range p.pool {
		if worker.ProcessingHost(host) {
			return worker
		}
	}

	if p.current >= len(p.pool) {
		p.current = p.current % len(p.pool)
	}
	return p.pool[p.current]
}

// Worker processing given jobs by jobs channel and sending responses back to results channel.
type Worker struct {
	jobs            chan entities.Job
	results         chan entities.Result
	processingHosts []entities.Host
}

// NewWorker returns new worker.
func NewWorker(jobsQueueSize int, results chan entities.Result) *Worker {
	return &Worker{
		jobs:    make(chan entities.Job, jobsQueueSize),
		results: results,
	}
}

// ProcessingHost returns true if worker is already processing job for given host.
func (w *Worker) ProcessingHost(host entities.Host) bool {
	return sort.Search(len(w.processingHosts), func(i int) bool { return w.processingHosts[i] == host }) < len(w.processingHosts)
}

// ProcessJobs processes job's channel using given clients configuration.
func (w *Worker) ProcessJobs(ctx context.Context, clientConfig clients.Clients) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-w.jobs:
			if !ok {
				return
			}

			var handler mode.OperationModeFunc
			var client clients.Client

			switch job.Kind {
			case mode.InitPublicKeySSHMode:
				client = clients.NewSSHClient(clientConfig.SSH)
				handler = mode.InitPublicKeySSH
			case mode.InitSecureAPIMode:
				client = clients.NewSSHClient(clientConfig.SSH)
				handler = mode.InitSecureAPI
			case mode.CustomSSHMode:
				client = clients.NewSSHClient(clientConfig.SSH)
				handler = mode.Custom
			case mode.CustomAPIMode:
				client = clients.NewMikrotikAPIClient(clientConfig.MikrotikAPI)
				handler = mode.Custom
			case mode.ChangePasswordMode:
				client = clients.NewMikrotikAPIClient(clientConfig.MikrotikAPI)
				handler = mode.ChangePassword
			default:
				w.results <- entities.Result{Host: job.Host, Job: job, Error: fmt.Errorf("invalid kind of job: %v", job)}
				continue
			}

			responses, err := handler(ctx, client, &job)
			w.results <- entities.Result{Host: job.Host, Job: job, Responses: responses, Error: err}
		}
	}
}
