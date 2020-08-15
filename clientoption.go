package fancyhttpclient

import (
	"time"

	"golang.org/x/time/rate"
)

// ClientOption apply custom options to the client
type ClientOption func(fhc *FancyHTTPClient) error

// WithDelay sets up the client to space out each requests sent
func WithDelay(delay time.Duration) ClientOption {
	return func(fhc *FancyHTTPClient) error {
		fhc.configureLimiterOnce.Do(func() {
			limiter := rate.NewLimiter(rate.Limit(time.Second/delay), 1)
			fhc.waiter = limiter
		})
		return nil
	}
}

// WithMaxConn sets up the client to limit its connection to a set number
func WithMaxConn(maxConn int) ClientOption {
	return func(fhc *FancyHTTPClient) error {
		fhc.maxConnection = maxConn
		return nil
	}
}

// WithLimiter configures the rate limiter on the client
func WithLimiter(l Waiter) ClientOption {
	return func(fhc *FancyHTTPClient) error {
		fhc.configureLimiterOnce.Do(func() {
			fhc.waiter = l
		})
		return nil
	}
}
