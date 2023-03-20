package httpcache

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"time"

	"github.com/gomodule/redigo/redis"

	"flamingo.me/flamingo/v3/framework/flamingo"
)

type (
	// RedisBackend implements the cache backend interface with a redis solution
	RedisBackend struct {
		cacheMetrics Metrics
		pool         *redis.Pool
		logger       flamingo.Logger
	}

	// RedisBackendFactory creates fully configured instances of Redis
	RedisBackendFactory struct {
		logger       flamingo.Logger
		frontendName string
		pool         *redis.Pool
		config       *RedisBackendConfig
	}

	// RedisBackendConfig holds the configuration values
	RedisBackendConfig struct {
		MaxIdle            int
		IdleTimeOutSeconds int
		Host               string
		Port               string
	}
)

const (
	tagPrefix   = "tag:"
	valuePrefix = "value:"
)

var (
	redisKeyRegex         = regexp.MustCompile(`[^a-zA-Z0-9]`)
	_             Backend = new(RedisBackend)
)

func init() {
	gob.Register(Entry{})
}

func finalizer(b *RedisBackend) {
	b.close()
}

// Inject Redis dependencies
func (f *RedisBackendFactory) Inject(logger flamingo.Logger) *RedisBackendFactory {
	f.logger = logger
	return f
}

// Build a new redis backend
func (f *RedisBackendFactory) Build() (Backend, error) {
	if f.config != nil {
		if f.config.IdleTimeOutSeconds <= 0 {
			return nil, errors.New("IdleTimeOut must be >0")
		}

		if f.config.Host == "" || f.config.Port == "" {
			return nil, errors.New("host and Port must set")
		}

		f.pool = &redis.Pool{
			MaxIdle:     f.config.MaxIdle,
			IdleTimeout: time.Second * time.Duration(f.config.IdleTimeOutSeconds),
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
			Dial: func() (redis.Conn, error) {
				return f.redisConnector(
					"tcp",
					fmt.Sprintf("%v:%v", f.config.Host, f.config.Port),
					"",
					0,
				)
			},
		}
	}

	b := &RedisBackend{
		pool:         f.pool,
		logger:       f.logger.WithField(flamingo.LogKeyCategory, "Redis"),
		cacheMetrics: NewCacheMetrics("redis", f.frontendName),
	}
	runtime.SetFinalizer(b, finalizer) // close all connections on destruction

	return b, nil
}

// SetFrontendName for redis cache metrics
func (f *RedisBackendFactory) SetFrontendName(frontendName string) *RedisBackendFactory {
	f.frontendName = frontendName
	return f
}

// SetConfig for redis
func (f *RedisBackendFactory) SetConfig(config RedisBackendConfig) *RedisBackendFactory {
	f.config = &config
	return f
}

// SetPool directly - use instead of SetConfig if desired
func (f *RedisBackendFactory) SetPool(pool *redis.Pool) *RedisBackendFactory {
	f.pool = pool
	return f
}

func (f *RedisBackendFactory) redisConnector(network, address, password string, db int) (redis.Conn, error) {
	c, err := redis.Dial(network, address)
	if err != nil {
		return nil, err
	}
	if password != "" {
		if _, err := c.Do("AUTH", password); err != nil {
			c.Close()
			return nil, err
		}
	}
	if db != 0 {
		if _, err := c.Do("SELECT", db); err != nil {
			c.Close()
			return nil, err
		}
	}
	return c, err
}

// Close ensures all redis connections are closed
func (b *RedisBackend) close() {
	b.pool.Close()
}

// createPrefixedKey creates a redis-compatible key
func (b *RedisBackend) createPrefixedKey(key string, prefix string) string {
	key = redisKeyRegex.ReplaceAllString(key, "-")
	return fmt.Sprintf("%v%v", prefix, key)
}

// Get a cache key
func (b *RedisBackend) Get(key string) (entry Entry, found bool) {
	conn := b.pool.Get()
	defer conn.Close()

	reply, err := conn.Do("GET", b.createPrefixedKey(key, valuePrefix))
	if err != nil {
		b.cacheMetrics.countError(fmt.Sprintf("%v", err))
		b.logger.Error(fmt.Sprintf("Error getting key '%v': %v", key, err))
		return Entry{}, false
	}
	if reply == nil {
		b.cacheMetrics.countMiss()
		b.logger.Info(fmt.Sprintf("Missed key: %v", key))
		return Entry{}, false
	}

	value, err := redis.Bytes(reply, err)
	if err != nil {
		b.cacheMetrics.countError("ByteConvertFailed")
		b.logger.Error(fmt.Sprintf("Error convert value to bytes of key '%v': %v", key, err))
		return Entry{}, false
	}

	redisEntry, err := b.decodeEntry(value)
	if err != nil {
		b.cacheMetrics.countError("DecodeFailed")
		b.logger.Error(fmt.Sprintf("Error decoding content of key '%v': %v", key, err))
		return Entry{}, false
	}

	b.cacheMetrics.countHit()
	return redisEntry, true
}

// Set a cache key
func (b *RedisBackend) Set(key string, entry Entry) error {
	conn := b.pool.Get()
	defer conn.Close()

	buffer, err := b.encodeEntry(entry)
	if err != nil {
		b.cacheMetrics.countError("EncodeFailed")
		b.logger.Error("Error encoding for key: ", key)
		return err
	}

	err = conn.Send(
		"SETEX",
		b.createPrefixedKey(key, valuePrefix),
		int(entry.Meta.GraceTime.Sub(time.Now().Round(time.Second))),
		buffer,
	)
	if err != nil {
		b.cacheMetrics.countError("SetFailed")
		b.logger.Error("Error setting key %v with timeout %v and buffer %v", key, entry.Meta.GraceTime, buffer)
		return err
	}

	for _, tag := range entry.Meta.Tags {
		err = conn.Send(
			"SADD",
			b.createPrefixedKey(tag, tagPrefix),
			b.createPrefixedKey(key, valuePrefix),
		)
		if err != nil {
			b.cacheMetrics.countError("SetTagFailed")
			b.logger.Error("Error setting tag: %v: %v", tag, key)
			return err
		}
	}

	return conn.Flush()
}

// Purge a cache key
func (b *RedisBackend) Purge(key string) error {
	conn := b.pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", b.createPrefixedKey(key, valuePrefix))
	return err
}

// PurgeTags purges all keys+tags by tag(s)
func (b *RedisBackend) PurgeTags(tags []string) error {
	conn := b.pool.Get()
	defer conn.Close()

	for _, tag := range tags {
		reply, err := conn.Do("SMEMBERS", b.createPrefixedKey(tag, tagPrefix))
		members, err := redis.Strings(reply, err)
		if err != nil {
			b.logger.Error(fmt.Sprintf("Failed SMEMBERS for tag '%v': %v", tag, err))
		}

		for _, member := range members {
			_, err = conn.Do("DEL", member)
			if err != nil {
				b.logger.Error(fmt.Sprintf("Failed DEL for key '%v': %v", member, err))
				return err
			}
		}

		_, err = conn.Do("DEL", fmt.Sprintf("%v", tag))
		if err != nil {
			b.logger.Error(fmt.Sprintf("Failed DEL for key '%v': %v", tag, err))
			return err
		}
	}
	return conn.Flush()
}

// Flush the whole cache
func (b *RedisBackend) Flush() error {
	conn := b.pool.Get()
	defer conn.Close()

	err := conn.Send("FLUSHALL")
	if err != nil {
		b.logger.Error(fmt.Sprintf("Failed purge all keys %v", err))
		return err
	}

	return conn.Flush()
}

func (b *RedisBackend) encodeEntry(entry Entry) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	err := gob.NewEncoder(buffer).Encode(entry)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

func (b *RedisBackend) decodeEntry(content []byte) (Entry, error) {
	buffer := bytes.NewBuffer(content)
	decoder := gob.NewDecoder(buffer)
	entry := new(Entry)
	err := decoder.Decode(entry)
	if err != nil {
		return *entry, err
	}

	return *entry, err
}
