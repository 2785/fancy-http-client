package fancyhttpclient

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkerPool(t *testing.T) {
	t.Run("50ms delay, 5 work", func(t *testing.T) {
		// 5 jobs, each job taking minimal time, 50ms delay, will cause the single worker to pick up the 5
		// works in sequence, waiting 50ms before each one, resulting in approximately 200ms of delay
		wp := NewWorkerPool(1, 50*time.Millisecond)
		wg := &sync.WaitGroup{}
		start := time.Now()
		for i := 0; i < 5; i++ {
			wg.Add(1)
			wp <- func() {
				wg.Done()
			}
		}
		wg.Wait()
		took := time.Since(start)
		t.Logf("took: %s", took)
		assert.True(t, took > 200*time.Millisecond && took < 210*time.Millisecond)
	})

	t.Run("50ms delay, 5 workers, 5 work taking 100ms each", func(t *testing.T) {
		// 5 workers, 50ms delay and 5 jobs taking 100ms each will cause 5 workers to pick up 5 jobs in sequence
		// with the delay, 200ms of wait time to start the job + 100ms to finish the last job started, should
		// take around 300ms
		wp := NewWorkerPool(5, 50*time.Millisecond)
		wg := &sync.WaitGroup{}
		start := time.Now()
		for i := 0; i < 5; i++ {
			wg.Add(1)
			wp <- func() {
				time.Sleep(100 * time.Millisecond)
				wg.Done()
			}
		}
		wg.Wait()
		took := time.Since(start)
		t.Logf("took: %s", took)
		assert.True(t, took > 300*time.Millisecond && took < 310*time.Millisecond)
	})

	t.Run("50ms delay, 2 workers, 5 work taking 100ms each", func(t *testing.T) {
		// 2 workers, 50ms delay and 5 jobs taking 100ms each will cause the following situation
		// 0ms: worker 1 pick up job 1
		// 50ms: worker 2 pick up job 2
		// 100ms: worker 1 finishes job 1, picks up job 3
		// 150ms: worker 2 finishes job 2, picks up job 4
		// 200ms: worker 1 finishes job 3, picks up job 5
		// 250ms: worker 2 finishes job 4
		// 300ms: worker 1 finishes job 5, all work done
		wp := NewWorkerPool(2, 50*time.Millisecond)
		wg := &sync.WaitGroup{}
		start := time.Now()
		for i := 0; i < 5; i++ {
			wg.Add(1)
			wp <- func() {
				time.Sleep(100 * time.Millisecond)
				wg.Done()
			}
		}
		wg.Wait()
		took := time.Since(start)
		t.Logf("took: %s", took)
		assert.True(t, took > 300*time.Millisecond && took < 310*time.Millisecond)
	})

	t.Run("50ms delay, 3 workers, 5 work taking 200ms each", func(t *testing.T) {
		// 2 workers, 50ms delay and 5 jobs taking 100ms each will cause the following situation
		// 0ms: worker 1 pick up job 1
		// 50ms: worker 2 pick up job 2
		// 100ms: worker 3 pick up job 3
		// 200ms: worker 1 finishes job 1, picks up job 4
		// 250ms: worker 2 finishes job 2, picks up job 5
		// 300ms: worker 3 finishes job 3
		// 400ms: worker 1 finishes job 4
		// 450ms: worker 2 finishes job 5, all work done
		wp := NewWorkerPool(3, 50*time.Millisecond)
		wg := &sync.WaitGroup{}
		start := time.Now()
		for i := 0; i < 5; i++ {
			wg.Add(1)
			wp <- func() {
				time.Sleep(200 * time.Millisecond)
				wg.Done()
			}
		}
		wg.Wait()
		took := time.Since(start)
		t.Logf("took: %s", took)
		assert.True(t, took > 450*time.Millisecond && took < 460*time.Millisecond)
	})

	t.Run("100ms delay, 3 workers, 5 work taking 50ms each", func(t *testing.T) {
		// at this point how many worker doesn't matter, as there's no work to parallelize with the set delay.
		// one / or many of the worker will pick up 5 jobs in sequence, waiting 100ms to grab the next one
		// will spend 400ms waiting, start last job at 400ms and finish at 450ms
		wp := NewWorkerPool(3, 100*time.Millisecond)
		wg := &sync.WaitGroup{}
		start := time.Now()
		for i := 0; i < 5; i++ {
			wg.Add(1)
			wp <- func() {
				time.Sleep(50 * time.Millisecond)
				wg.Done()
			}
		}
		wg.Wait()
		took := time.Since(start)
		t.Logf("took: %s", took)
		assert.True(t, took > 450*time.Millisecond && took < 460*time.Millisecond)
	})
}

func TestNewWorkerPool(t *testing.T) {
	// should panic if given 0 worker to work with, just can't do that :)
	assert.Panics(t, func() {
		NewWorkerPool(0, 0)
	})
}
