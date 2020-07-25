package fancyhttpclient

import "time"

// ClientOption apply custom options to the client
type ClientOption func(fhc *FancyHTTPClient)

// WithDelay sets up the client to space out each requests sent
func WithDelay(delay time.Duration) ClientOption {
	return func(fhc *FancyHTTPClient) {
		fhc.delay = delay
	}
}

// WithMaxConn sets up the client to limit its connection to a set number
func WithMaxConn(maxConn int) ClientOption {
	return func(fhc *FancyHTTPClient) {
		fhc.maxConnection = maxConn
	}
}
