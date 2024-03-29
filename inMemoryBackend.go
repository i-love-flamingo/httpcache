package httpcache

import (
	"fmt"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

const defaultLurkerPeriod = 1 * time.Minute

type (
	// MemoryBackend implements the cache backend interface with an "in memory" solution
	MemoryBackend struct {
		cacheMetrics Metrics
		pool         *lru.TwoQueueCache[string, inMemoryCacheEntry]
		lurkerPeriod time.Duration
	}

	// MemoryBackendConfig config
	MemoryBackendConfig struct {
		Size int
	}

	// InMemoryBackendFactory factory
	InMemoryBackendFactory struct {
		config       MemoryBackendConfig
		frontendName string
		lurkerPeriod time.Duration
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

// SetLurkerPeriod sets the timeframe how often expired cache entries should be checked/cleaned up, if 0 is provided the default period of 1 minute is taken
func (f *InMemoryBackendFactory) SetLurkerPeriod(period time.Duration) *InMemoryBackendFactory {
	f.lurkerPeriod = period
	return f
}

// SetFrontendName used in Metrics
func (f *InMemoryBackendFactory) SetFrontendName(frontendName string) *InMemoryBackendFactory {
	f.frontendName = frontendName
	return f
}

// Build the instance
func (f *InMemoryBackendFactory) Build() (Backend, error) {
	cache, _ := lru.New2Q[string, inMemoryCacheEntry](f.config.Size)

	lurkerPeriod := defaultLurkerPeriod
	if f.lurkerPeriod > 0 {
		lurkerPeriod = f.lurkerPeriod
	}

	memoryBackend := &MemoryBackend{
		pool:         cache,
		cacheMetrics: NewCacheMetrics("memory", f.frontendName),
		lurkerPeriod: lurkerPeriod,
	}

	go memoryBackend.lurker()

	return memoryBackend, nil
}

// SetSize creates a new underlying cache of the given size
func (m *MemoryBackend) SetSize(size int) error {
	cache, err := lru.New2Q[string, inMemoryCacheEntry](size)
	if err != nil {
		return fmt.Errorf("lru cant create new TwoQueueCache: %w", err)
	}

	m.pool = cache

	return nil
}

// Get tries to get an object from cache
func (m *MemoryBackend) Get(key string) (Entry, bool) {
	entry, found := m.pool.Get(key)
	if !found {
		m.cacheMetrics.countMiss()
		return Entry{}, false
	}

	m.cacheMetrics.countHit()

	data, ok := entry.data.(Entry)
	if !ok {
		return Entry{}, false
	}

	return data, true
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
	for range time.Tick(m.lurkerPeriod) {
		m.cacheMetrics.recordEntries(int64(m.pool.Len()))

		for _, key := range m.pool.Keys() {
			entry, found := m.pool.Peek(key)
			if found && entry.valid.Before(time.Now()) {
				m.pool.Remove(key)
				break
			}
		}
	}
}
