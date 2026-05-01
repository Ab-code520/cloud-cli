package utils

import (
	"context"
	"sync"
)

// Pool manages a pool of concurrent workers with context cancellation.
type Pool struct {
	sem    chan struct{}
	wg     sync.WaitGroup
	mu     sync.Mutex
	err    error
	ctx    context.Context
	cancel context.CancelFunc
}

// NewPool creates a new worker pool.
func NewPool(concurrency int) *Pool {
	if concurrency < 1 {
		concurrency = 1
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		sem:    make(chan struct{}, concurrency),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Go runs the function in the worker pool.
// Go submits a task to the pool. Uses select to avoid blocking when context is cancelled.
func (p *Pool) Go(fn func(ctx context.Context) error) {
	// Fix #1: Context-aware submission to prevent CPU spinning during fast-fail
	select {
	case <-p.ctx.Done():
		// Already cancelled, discard task immediately
		return
	case p.sem <- struct{}{}:
		// Acquired semaphore
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer func() { <-p.sem }()

		// Double-check context in case cancelled while waiting for semaphore
		if p.ctx.Err() != nil {
			return
		}

		if err := fn(p.ctx); err != nil {
			p.mu.Lock()
			if p.err == nil {
				p.err = err
				p.cancel()
			}
			p.mu.Unlock()
		}
	}()
}

// Wait blocks until all workers complete.
func (p *Pool) Wait() error {
	p.wg.Wait()
	p.cancel() // Ensure cleanup
	return p.err
}
