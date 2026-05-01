package utils

import (
	"context"
	"sync"
	"time"
)

// TokenBucket 令牌桶限流器
type TokenBucket struct {
	tokens   chan struct{}
	interval time.Duration
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(maxConcurrent int, interval time.Duration) *TokenBucket {
	return &TokenBucket{
		tokens:   make(chan struct{}, maxConcurrent),
		interval: interval,
	}
}

// Acquire 获取令牌
func (tb *TokenBucket) Acquire(ctx context.Context) error {
	select {
	case tb.tokens <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release 释放令牌
func (tb *TokenBucket) Release() {
	select {
	case <-tb.tokens:
	default:
	}
}

// WorkerPool 协程池
type WorkerPool struct {
	sem     chan struct{}
	wg      sync.WaitGroup
	results []chan interface{}
}

// NewWorkerPool 创建协程池
func NewWorkerPool(maxWorkers int) *WorkerPool {
	return &WorkerPool{
		sem: make(chan struct{}, maxWorkers),
	}
}

// Go 提交任务到池中
func (wp *WorkerPool) Go(fn func()) {
	wp.wg.Add(1)
	wp.sem <- struct{}{}
	go func() {
		defer wp.wg.Done()
		defer func() { <-wp.sem }()
		fn()
	}()
}

// Wait 等待所有任务完成
func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}
