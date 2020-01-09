# HTTPCache module
The httpcache module provides an easy interface to cache simple http results in flamingo.

The basic concept is, that there is a so called "cache frontend" - that offers an interface to cache certain types, 
and a "cache backend" that takes care about storing(persisting) the cache entry.

## Caching HTTP responses from APIs

A typical use case is, to cache responses from (slow) APIs that you need to call.

First define an injection to get a Frontend cache injected:

```go
MyApiClient struct {
      Cache  *httpcache.Frontend
}
```

## Cache backends

Currently there are the following backends available:

### inMemoryCache

Caches in memory - and therefore is a very fast cache.

It is base on the LRU-Strategy witch drops least used entries. For this reason the cache will be no overcommit your memory and will automicly fit the need of your current traffic.

### redisBackend

Is using [redis](https://redis.io/) as an shared inMemory cache.
Since all cache-fetched has an overhead to the inMemoryBackend, the redis is a little slower.
The benefit of redis is the shared storage an the high efficiency in reading and writing keys. especialy if you need scale fast horizonaly, it helps to keep your backend-systems healthy.

Be ware of using redis (or any other shared cache backend) as an single backend, because of network latency. (have a look at the twoLevelBackend)


### twoLevelBackend

The twoLevelBackend was introduced to get the benefit of the extreme fast inMemoryBackend and a shared backend.
Using the inMemoryBackend in combination with an shared backend, gives you blazing fast responses and helps you to protect your backend in case of fast scaleout-scenarios.

