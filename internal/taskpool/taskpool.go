// Package taskpool provides a pool that receives tasks and executes them with a concurrency limit.
package taskpool

import (
	"sync"
)

type task func()

type worker struct{}

func (w *worker) do(ch <-chan task, wg *sync.WaitGroup) {
	for t := range ch {
		t()
		wg.Done()
	}
}

// Pool represents a pool of tasks, which will be executed concurrently with a limit.
type Pool struct {
	wg      sync.WaitGroup
	tasksCh chan task
	workers []worker
}

// New creates a new pool.
func New(maxConcurrency int) *Pool {
	workers := []worker{}
	for i := 0; i < maxConcurrency; i++ {
		w := worker{}
		workers = append(workers, w)
	}

	p := &Pool{
		wg:      sync.WaitGroup{},
		tasksCh: make(chan task, maxConcurrency),
		workers: workers,
	}

	for i := range workers {
		go workers[i].do(p.tasksCh, &p.wg)
	}

	return p
}

// Add lets you add a task to the Pool.
func (p *Pool) Add(task task) {
	p.wg.Add(1)
	p.tasksCh <- task
}

// Wait waits until all tasks added to the Pool have finished.
func (p *Pool) Wait() {
	p.wg.Wait()
	close(p.tasksCh)
}
