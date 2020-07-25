package fancyhttpclient

import (
	"net/http"
	"sync"
	"time"
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

	c.workerPool = func() WorkerPool {
		if c.maxConnection == 0 {
			return NewWorkerPool(1, c.delay)
		}
		return NewWorkerPool(c.maxConnection, c.delay)
	}()

	return c
}

// FancyHTTPClient is a custom HTTP client that is able to handle
type FancyHTTPClient struct {
	httpClient    Doer
	delay         time.Duration
	maxConnection int
	workerPool    WorkerPool
}

// Do fires off one single request
func (c *FancyHTTPClient) Do(req *http.Request) (*http.Response, error) {
	resChan, errChan := make(chan *http.Response), make(chan error)
	c.workerPool <- func() {
		res, err := c.httpClient.Do(req)
		resChan <- res
		errChan <- err
		close(resChan)
		close(errChan)
		return
	}
	return <-resChan, <-errChan
}

// DoBunch fires off a bunch of requests with the configured delay and returns all, reserving the order of slice passed in
func (c *FancyHTTPClient) DoBunch(reqs []*http.Request) []*ResponseGetter {
	responses := make([]*ResponseGetter, len(reqs))

	wg := &sync.WaitGroup{}
	for i, r := range reqs {
		ind := i
		req := r
		wg.Add(1)
		c.workerPool <- func() {
			res, err := c.httpClient.Do(req)
			responses[ind] = &ResponseGetter{res, err}
			wg.Done()
			return
		}
	}
	wg.Wait()
	return responses
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
