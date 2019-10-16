package service

import (
	"sync"

	"github.com/migotom/mt-bulk/internal/entities"
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

	p.current++
	if p.current >= len(p.pool) {
		p.current = p.current % len(p.pool)
	}
	return p.pool[p.current]
}
