package httpcache

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"time"

	"flamingo.me/flamingo/v3/core/healthcheck/domain/healthcheck"
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
	_ Backend            = new(RedisBackend)
	_ healthcheck.Status = new(RedisBackend)

	redisKeyRegex = regexp.MustCompile(`[^a-zA-Z0-9]`)

	ErrInvalidRedisConfig = errors.New("invalid redis config")
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
			return nil, fmt.Errorf("IdleTimeOut must be >0: %w", ErrInvalidRedisConfig)
		}

		if f.config.Host == "" || f.config.Port == "" {
			return nil, fmt.Errorf("host and port must set: %w", ErrInvalidRedisConfig)
		}

		f.pool = &redis.Pool{
			MaxIdle:     f.config.MaxIdle,
			IdleTimeout: time.Second * time.Duration(f.config.IdleTimeOutSeconds),
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return fmt.Errorf("redis PING failed: %w", err)
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

	redisBackend := &RedisBackend{
		pool:         f.pool,
		logger:       f.logger.WithField(flamingo.LogKeyCategory, "Redis"),
		cacheMetrics: NewCacheMetrics("redis", f.frontendName),
	}
	runtime.SetFinalizer(redisBackend, finalizer) // close all connections on destruction

	return redisBackend, nil
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

func (f *RedisBackendFactory) redisConnector(network, address, password string, database int) (redis.Conn, error) {
	conn, err := redis.Dial(network, address)
	if err != nil {
		return nil, fmt.Errorf("redis dial error: %w", err)
	}

	if password != "" {
		if _, err := conn.Do("AUTH", password); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("redis auth error: %w", err)
		}
	}

	if database != 0 {
		if _, err := conn.Do("SELECT", database); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("redis select db error: %w", err)
		}
	}

	return conn, nil
}

// Close ensures all redis connections are closed
func (b *RedisBackend) close() {
	_ = b.pool.Close()
}

// createPrefixedKey creates a redis-compatible key
func (b *RedisBackend) createPrefixedKey(key string, prefix string) string {
	key = redisKeyRegex.ReplaceAllString(key, "-")
	return fmt.Sprintf("%v%v", prefix, key)
}

// Get a cache key
func (b *RedisBackend) Get(key string) (entry Entry, found bool) {
	conn := b.pool.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	reply, err := conn.Do("GET", b.createPrefixedKey(key, valuePrefix))
	if err != nil {
		b.cacheMetrics.countError(fmt.Sprintf("%v", err))
		b.logger.Error(fmt.Sprintf("Error getting key '%v': %v", key, err))

		return Entry{}, false
	}

	if reply == nil {
		b.cacheMetrics.countMiss()

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
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	buffer, err := b.encodeEntry(entry)
	if err != nil {
		b.cacheMetrics.countError("EncodeFailed")
		b.logger.Error(fmt.Sprintf("Error encoding entry for key: %q", key))

		return err
	}

	err = conn.Send(
		"SET",
		b.createPrefixedKey(key, valuePrefix),
		buffer,
		"PX",
		time.Until(entry.Meta.GraceTime).Round(time.Millisecond).Milliseconds(),
	)
	if err != nil {
		b.cacheMetrics.countError("SetFailed")
		b.logger.Error(fmt.Sprintf("Error setting key %q with timeout %v and buffer %v", key, entry.Meta.GraceTime, buffer))

		return fmt.Errorf("redis SET PX failed: %w", err)
	}

	for _, tag := range entry.Meta.Tags {
		err = conn.Send(
			"SADD",
			b.createPrefixedKey(tag, tagPrefix),
			b.createPrefixedKey(key, valuePrefix),
		)
		if err != nil {
			b.cacheMetrics.countError("SetTagFailed")
			b.logger.Error(fmt.Sprintf("Error setting tag: %q on key %q", tag, key))

			return fmt.Errorf("redis SADD failed: %w", err)
		}
	}

	err = conn.Flush()
	if err != nil {
		return fmt.Errorf("redis flush failed: %w", err)
	}

	return nil
}

// Purge a cache key
func (b *RedisBackend) Purge(key string) error {
	conn := b.pool.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	_, err := conn.Do("DEL", b.createPrefixedKey(key, valuePrefix))
	if err != nil {
		return fmt.Errorf("redis DEL failed: %w", err)
	}

	return nil
}

// PurgeTags purges all keys+tags by tag(s)
func (b *RedisBackend) PurgeTags(tags []string) error {
	conn := b.pool.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

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

				return fmt.Errorf("redis DEL failed for key %q: %w", member, err)
			}
		}

		_, err = conn.Do("DEL", fmt.Sprintf("%v", tag))
		if err != nil {
			b.logger.Error(fmt.Sprintf("Failed DEL for key '%v': %v", tag, err))

			return fmt.Errorf("redis DEL failed for key %q: %w", tag, err)
		}
	}

	err := conn.Flush()
	if err != nil {
		return fmt.Errorf("redis flush failed: %w", err)
	}

	return nil
}

// Flush the whole cache
func (b *RedisBackend) Flush() error {
	conn := b.pool.Get()
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	err := conn.Send("FLUSHALL")
	if err != nil {
		b.logger.Error(fmt.Sprintf("Failed purge all keys %v", err))

		return fmt.Errorf("redis FLUSHALL failed: %w", err)
	}

	err = conn.Flush()
	if err != nil {
		return fmt.Errorf("redis flush failed: %w", err)
	}

	return nil
}

func (b *RedisBackend) encodeEntry(entry Entry) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)

	err := gob.NewEncoder(buffer).Encode(entry)
	if err != nil {
		return nil, fmt.Errorf("gob encode failed: %w", err)
	}

	return buffer, nil
}

func (b *RedisBackend) decodeEntry(content []byte) (Entry, error) {
	buffer := bytes.NewBuffer(content)
	decoder := gob.NewDecoder(buffer)
	entry := new(Entry)

	err := decoder.Decode(entry)
	if err != nil {
		return *entry, fmt.Errorf("gob decode failed: %w", err)
	}

	return *entry, nil
}

// Status checks the health of the used redis instance
func (b *RedisBackend) Status() (bool, string) {
	_, err := b.pool.Get().Do("PING")
	if err != nil {
		return false, fmt.Sprintf("redis PING failed: %q", err.Error())
	}

	return true, ""
}
