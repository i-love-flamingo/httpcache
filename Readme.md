# HTTPCache module

[![Go Report Card](https://goreportcard.com/badge/github.com/i-love-flamingo/httpcache)](https://goreportcard.com/report/github.com/i-love-flamingo/httpcache)
[![GoDoc](https://godoc.org/github.com/i-love-flamingo/httpcache?status.svg)](https://godoc.org/github.com/i-love-flamingo/httpcache)
[![Tests](https://github.com/i-love-flamingo/httpcache/workflows/Tests/badge.svg?branch=master)](https://github.com/i-love-flamingo/httpcache/actions?query=branch%3Amaster+workflow%3ATests)

The HTTPCache module provides an easy interface to cache simple HTTP results in Flamingo.

The basic concept is, that there is a so-called "cache frontend" - that offers an interface to cache certain types, 
and a "cache backend" that takes care about storing(persisting) the cache entry.

## Caching HTTP responses from APIs

A typical use case is, to cache responses from (slow) APIs that you need to call.

First, add the dependency to your project:
```bash
go get flamingo.me/httpcache
```

For an easy start the module ships with a cache frontend factory, we'll configure a frontend that relies on an in-memory backend.
The name of the frontend will be `myServiceCache` and the cache can store up to 50 entries before dropping the least frequently used.

Add this to your config.yaml:
```yaml
httpcache:
  frontendFactory:
    myServiceCache:
      backendType: memory
      memory:
        size: 50 // limit of entries
```

__For a list of all supported cache backends see [the cache backends section](#cache-backends) below.__

Afterward you can use the cache frontend in your API client:

```go
package api

import (
	"context"
	"net/http"
	"time"

	"flamingo.me/httpcache"
)

type MyApiClient struct {
	Cache *httpcache.Frontend `inject:"myServiceCache"`
}

type Result struct{}

func (m *MyApiClient) Operation(ctx context.Context) (*Result, error) {
	cacheEntry, err := m.Cache.Get(ctx, "operation-cache-key", m.doApiCall(ctx, "https://example.com/v1/operation"))
	if err != nil {
		return nil, err
	}

	// unmarshall cacheEntry.Body map to Result struct
	_ = cacheEntry.Body

	return nil, nil
}

func (m *MyApiClient) doApiCall(ctx context.Context, endpoint string) httpcache.HTTPLoader {
	return func(ctx context.Context) (httpcache.Entry, error) {
		// grab http client with timeout
		// call endpoint
		_ = endpoint

		return httpcache.Entry{
			Meta: httpcache.Meta{
				LifeTime:  time.Now().Add(5 * time.Minute),
				GraceTime: time.Now().Add(10 * time.Minute),
				Tags:      nil,
			},
			Body:       []byte{}, // API Response that should be cached
			StatusCode: http.StatusOK,
		}, nil
	}
}

```

## Cache backends

Currently, there are the following backends available:

### In memory

`backendType: memory`

Caches in memory - and therefore is a very fast cache.

It is base on the LRU-Strategy witch drops the least used entries. For this reason the cache will be no over-commit your memory and will atomically fit the need of your current traffic.

Example config:
```yaml
httpcache:
  frontendFactory:
    myServiceCache:
      backendType: memory
      memory:
        size: 200 // limit of entries
```

### Redis

`backendType: redis`

Is using [redis](https://redis.io/) as a shared inMemory cache.
Since all cache-fetched has an overhead to the inMemoryBackend, the redis is a little slower.
The benefit of redis is the shared storage and the high efficiency in reading and writing keys. Especially if you need scale fast horizontally, it helps to keep your backend-systems healthy.

Be aware of using redis (or any other shared cache backend) as a single backend, because of network latency. (have a look at the twoLevelBackend)

```yaml
httpcache:
  frontendFactory:
    myServiceCache:
      backendType: redis
      redis:
        host: '%%ENV:REDISHOST%%localhost%%'
        port: '6379'
```

### Two Level

`backendType: twolevel`

The two level backend was introduced to get the benefit of the extreme fast memory backend and a shared backend.
Using the inMemoryBackend in combination with a shared backend, gives you blazing fast responses and helps you to protect your backend in case of fast scaleout-scenarios.

Example config using the frontend cache factory:
```yaml
httpcache:
  frontendFactory:
    myServiceCache:
      backendType: twolevel
      twolevel:
        first:
          backendType: memory
          memory:
            size: 200
        second:
          backendType: redis
          redis:
            # Close connections after remaining idle for this duration. If the value
            # is zero, then idle connections are not closed. Applications should set
            # the timeout to a value less than the server's timeout.
            idleTimeOutSeconds: 60
            host: '%%ENV:REDISHOST%%localhost%%'
            port: '6379'
            maxIdle: 8
```

### Implement custom cache backend

If you are missing a cache backend feel free to open a issue or pull request.
It's of course possible to implement a custom cache backend in your project, see example below:

```go
package cache_backend

import (
	"flamingo.me/dingo"
	"flamingo.me/httpcache"
)

type Module struct {
	provider httpcache.FrontendProvider
}

type CustomBackend struct {
	// implement logic
}

var _ httpcache.Backend = new(CustomBackend)

// Inject dependencies
func (m *Module) Inject(
	provider httpcache.FrontendProvider,
) *Module {
	m.provider = provider
	return m
}

// Configure DI
func (m *Module) Configure(injector *dingo.Injector) {
	frontend := m.provider()
	frontend.SetBackend(&CustomBackend{})

	injector.Bind((*httpcache.Frontend)(nil)).AnnotatedWith("myServiceWithCustomBackend").ToInstance(frontend)
}
```