# Fancy HTTP Client

Golang HTTP Client that satisfies the `Doer` interface, i.e.

```Go
... interface {
    Do(*http.Request) (*http.Response, error)
}
```

With configurable throttling and / or connection limiting. 

## Quick Start

Base client with no throttling and connection limit, this is effectively the same as using the base client itself. 

```Go
baseHC := &http.Client{
    Timeout: 30 * time.Second
}

fhc := fancyhttpclient.New(baseHC)
```

Client with delay between outgoing requests, this will space out outgoing requests on a 50ms interval. The delay is applied across threads if threads share the same instance of fhc. 

```Go
fhc := fancyhttpclient.New(baseHC, fancyhttpclient.WithDelay(50*time.Millisecond))
```

Client with limit to the max number of connections it has to the server, will limit to max 5 connections

```Go
fhc := fancyhttpclient.New(baseHC, fancyhttpclient.WithMaxCoon(5))
```

Or both! This will limit max connection to 5 and space out the outgoing requests by 50ms

```Go
fhc := fancyhttpclient.New(
    baseHC, 
    fancyhttpclient.WithDelay(50*time.Millisecond),
    fancyhttpclient.WithMaxCoon(5),
)
```

It can then be passed to multiple goroutines that each performs individual `Do()`s, or

```Go
reqs := []*http.Request{bunch, of, requests}
responsers, _ := fhc.DoBunch(reqs)
res0, err0 := responsers[0].Response()
```