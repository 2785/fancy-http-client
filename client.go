package fancyhttpclient

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gammazero/workerpool"
)

// Doer performs an HTTP request
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// New returns an instance of FancyHTTPClient according to the delay and timeout provided
func New(hc Doer, options ...ClientOption) *FancyHTTPClient {
	c := &FancyHTTPClient{
		httpClient: hc,
	}

	for _, opt := range options {
		opt(c)
	}

	c.workerPool = workerpool.New(c.maxConnection)

	c.doneChan = make(chan struct{})

	c.mut = &sync.RWMutex{}

	c.delayChan = make(chan struct{})
	go func() {
		c.delayChan <- struct{}{}
		for {
			time.Sleep(c.delay)
			select {
			case c.delayChan <- struct{}{}:
			case <-c.doneChan:
				close(c.delayChan)
				return
			}
		}
	}()

	return c
}

// FancyHTTPClient is a custom HTTP client that is able to handle
type FancyHTTPClient struct {
	httpClient    Doer
	delay         time.Duration
	maxConnection int
	workerPool    *workerpool.WorkerPool
	delayChan     chan struct{}
	doneChan      chan struct{}
	mut           *sync.RWMutex
}

// Do fires off one single request
func (c *FancyHTTPClient) Do(req *http.Request) (*http.Response, error) {
	resChan, errChan := make(chan *http.Response), make(chan error)
	c.mut.RLock()
	if c.workerPool.Stopped() {
		return nil, errors.New("client has been terminated, cannot send request")
	}
	c.workerPool.Submit(func() {
		<-c.delayChan
		res, err := c.httpClient.Do(req)
		resChan <- res
		errChan <- err
		close(resChan)
		close(errChan)
		return
	})
	res, err := <-resChan, <-errChan
	c.mut.RUnlock()
	return res, err
}

// DoBunch fires off a bunch of requests with the configured delay and returns all, reserving the order of slice passed in
func (c *FancyHTTPClient) DoBunch(reqs []*http.Request) ([]*ResponseGetter, error) {
	responses := make([]*ResponseGetter, len(reqs))
	c.mut.RLock()
	if c.workerPool.Stopped() {
		return nil, errors.New("client has been terminated, cannot send request")
	}
	wg := &sync.WaitGroup{}
	for i, r := range reqs {
		ind := i
		req := r
		wg.Add(1)
		c.workerPool.Submit(func() {
			<-c.delayChan
			res, err := c.httpClient.Do(req)
			responses[ind] = &ResponseGetter{res, err}
			wg.Done()
			return
		})
	}
	wg.Wait()
	c.mut.RUnlock()
	return responses, nil
}

// Destroy waits for existing work to finish and stops the workerpool, function will only return after all work has been done
func (c *FancyHTTPClient) Destroy() {
	c.mut.Lock()
	c.workerPool.StopWait()
	c.doneChan <- struct{}{}
	close(c.doneChan)
	c.mut.Unlock()
}

// ResponseGetter contains either an http response or an error
type ResponseGetter struct {
	res *http.Response
	err error
}

// Response returns the response or error wrapped in the ResponseGetter
func (rg *ResponseGetter) Response() (*http.Response, error) {
	return rg.res, rg.err
}
