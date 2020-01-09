package httpcache

import (
	"time"

	lru "github.com/hashicorp/golang-lru"
)

const lurkerPeriod = 1 * time.Minute

type (
	// MemoryBackend implements the cache backend interface with an "in memory" solution
	MemoryBackend struct {
		cacheMetrics Metrics
		pool         *lru.TwoQueueCache
	}

	// MemoryBackendConfig config
	MemoryBackendConfig struct {
		Size int
	}

	// InMemoryBackendFactory factory
	InMemoryBackendFactory struct {
		config       MemoryBackendConfig
		frontendName string
	}

	inMemoryCacheEntry struct {
		valid time.Time
		data  interface{}
	}
)

var _ Backend = new(MemoryBackend)

// SetConfig for factory
func (f *InMemoryBackendFactory) SetConfig(config MemoryBackendConfig) *InMemoryBackendFactory {
	f.config = config
	return f
}

// SetFrontendName used in Metrics
func (f *InMemoryBackendFactory) SetFrontendName(frontendName string) *InMemoryBackendFactory {
	f.frontendName = frontendName
	return f
}

// Build the instance
func (f *InMemoryBackendFactory) Build() (Backend, error) {
	cache, _ := lru.New2Q(f.config.Size)

	m := &MemoryBackend{
		pool:         cache,
		cacheMetrics: NewCacheMetrics("memory", f.frontendName),
	}
	go m.lurker()
	return m, nil
}

// SetSize creates a new underlying cache of the given size
func (m *MemoryBackend) SetSize(size int) error {
	cache, err := lru.New2Q(size)
	if err != nil {
		return err
	}
	m.pool = cache
	return nil
}

// Get tries to get an object from cache
func (m *MemoryBackend) Get(key string) (Entry, bool) {
	entry, ok := m.pool.Get(key)
	if !ok {
		m.cacheMetrics.countMiss()
		return Entry{}, false
	}
	m.cacheMetrics.countHit()
	return entry.(inMemoryCacheEntry).data.(Entry), ok
}

// Set a cache entry with a key
func (m *MemoryBackend) Set(key string, entry Entry) error {
	m.pool.Add(key, inMemoryCacheEntry{
		data:  entry,
		valid: entry.Meta.GraceTime,
	})

	return nil
}

// Purge a cache key
func (m *MemoryBackend) Purge(key string) error {
	m.pool.Remove(key)

	return nil
}

// Flush purges all entries in the cache
func (m *MemoryBackend) Flush() error {
	m.pool.Purge()

	return nil
}

func (m *MemoryBackend) lurker() {
	for range time.Tick(lurkerPeriod) {
		for _, key := range m.pool.Keys() {
			item, ok := m.pool.Peek(key)
			if ok && item.(inMemoryCacheEntry).valid.Before(time.Now()) {
				m.pool.Remove(key)
				break
			}
		}
	}
}
