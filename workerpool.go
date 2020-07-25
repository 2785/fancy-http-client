package fancyhttpclient

import (
	"time"
)

// WorkerPool sends work to a number of goroutines to perform
type WorkerPool chan func()

// NewWorkerPool generates a worker pool with a given number of goroutines that do work
func NewWorkerPool(count int, delay time.Duration) WorkerPool {
	if count == 0 {
		panic("nah fam ya can't make a workerpool with 0 workers")
	}

	canStart := make(chan struct{})
	started := make(chan struct{})

	go func() {
		canStart <- struct{}{}
		for {
			<-started
			time.Sleep(delay)
			canStart <- struct{}{}
		}
	}()

	pool := make(chan func())

	for i := 0; i < count; i++ {
		go func() {
			for work := range pool {
				// do work
				<-canStart
				started <- struct{}{}
				work()
			}
		}()
	}

	return pool
}
