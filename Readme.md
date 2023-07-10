# HTTPCache module

[![Go Report Card](https://goreportcard.com/badge/github.com/i-love-flamingo/httpcache)](https://goreportcard.com/report/github.com/i-love-flamingo/httpcache)
[![GoDoc](https://godoc.org/github.com/i-love-flamingo/httpcache?status.svg)](https://godoc.org/github.com/i-love-flamingo/httpcache)
[![Tests](https://github.com/i-love-flamingo/httpcache/workflows/Tests/badge.svg?branch=master)](https://github.com/i-love-flamingo/httpcache/actions?query=branch%3Amaster+workflow%3ATests)

The httpcache module provides an easy interface to cache simple http results in Flamingo.

The basic concept is, that there is a so-called "cache frontend" - that offers an interface to cache certain types, 
and a "cache backend" that takes care about storing(persisting) the cache entry.

## Caching HTTP responses from APIs

A typical use case is, to cache responses from (slow) APIs that you need to call.

First define an injection to get a Frontend cache injected:

```go
type MyApiClient struct {
      Cache  *httpcache.Frontend  `inject:"myservice"`
}
```

We use annotation to be able to individually configure the requested Cache. So our binding may look like:

```go
injector.Bind((*Frontend)(nil)).AnnotatedWith("myservice").ToInstance(&Frontend{})
```

## Cache backends

Currently, there are the following backends available:

### memory

Caches in memory - and therefore is a very fast cache.

It is base on the LRU-Strategy witch drops least used entries. For this reason the cache will be no over-commit your memory and will atomically fit the need of your current traffic.

### redis

Is using [redis](https://redis.io/) as a shared inMemory cache.
Since all cache-fetched has an overhead to the inMemoryBackend, the redis is a little slower.
The benefit of redis is the shared storage and the high efficiency in reading and writing keys. Especially if you need scale fast horizontally, it helps to keep your backend-systems healthy.

Be aware of using redis (or any other shared cache backend) as a single backend, because of network latency. (have a look at the twoLevelBackend)


### twoLevel

The twoLevelBackend was introduced to get the benefit of the extreme fast memory backend and a shared backend.
Using the inMemoryBackend in combination with a shared backend, gives you blazing fast responses and helps you to protect your backend in case of fast scaleout-scenarios.


## Using Cache Factory

To automatically bind a Frontend for a cache "CACHENAME" you can use the cache.Module and configure the factory:

```yaml
httpcache:
  frontendFactory:
    CACHENAME:
      backendType: memory
      memory:
        size: 200
```

Or to configure a TwoLevel Cache:

```yaml
httpcache:
  frontendFactory:
    CACHENAME:
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

See `CueConfig` function in module.go for the complete config specification.

Afterward you can use the cache frontend:

```go
type MyApiClient struct {
      Cache  *httpcache.Frontend  `inject:"CACHENAME"`
}
```