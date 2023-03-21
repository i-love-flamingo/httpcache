package httpcache

import (
	"errors"
	"fmt"

	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3/framework/config"
)

var ErrRedisConfig = errors.New("redis config not complete")
var ErrMemoryConfig = errors.New("memory config not complete")
var ErrTwoLevelConfig = errors.New("twolevel config not complete")
var ErrInvalidBackend = errors.New("invalid backend supplied")

type (
	// FrontendFactory that can be used to build caches
	FrontendFactory struct {
		provider               FrontendProvider
		redisBackendFactory    *RedisBackendFactory
		inMemoryBackendFactory *InMemoryBackendFactory
		twoLevelBackendFactory *TwoLevelBackendFactory
		cacheConfig            FactoryConfig
	}

	// FactoryConfig typed configuration used to build Caches by the factory
	FactoryConfig map[string]BackendConfig

	// BackendConfig typed configuration used to build BackendCaches by the factory
	BackendConfig struct {
		BackendType string
		Memory      *MemoryBackendConfig
		Redis       *RedisBackendConfig
		Twolevel    *struct {
			First  *BackendConfig
			Second *BackendConfig
		}
	}

	// FrontendProvider - Dingo Provider func
	FrontendProvider func() *Frontend
)

// Inject for dependencies
func (f *FrontendFactory) Inject(
	provider FrontendProvider,
	redisBackendFactory *RedisBackendFactory,
	inMemoryBackendFactory *InMemoryBackendFactory,
	twoLevelBackendFactory *TwoLevelBackendFactory,
	cfg *struct {
		CacheConfig config.Map `inject:"config:httpcache.frontendFactory,optional"`
	},
) *FrontendFactory {
	f.provider = provider
	f.inMemoryBackendFactory = inMemoryBackendFactory
	f.redisBackendFactory = redisBackendFactory
	f.twoLevelBackendFactory = twoLevelBackendFactory

	if cfg != nil {
		var cacheConfig FactoryConfig

		err := cfg.CacheConfig.MapInto(&cacheConfig)
		if err != nil {
			panic(err)
		}

		f.cacheConfig = cacheConfig
	}

	return f
}

// BindConfiguredCaches creates annotated bindings from the cache configuration
func (f *FrontendFactory) BindConfiguredCaches(injector *dingo.Injector) error {
	for cacheName, cfg := range f.cacheConfig {
		backend, err := f.BuildBackend(cfg, cacheName)
		if err != nil {
			return err
		}

		injector.Bind((*Frontend)(nil)).AnnotatedWith(cacheName).ToInstance(f.BuildWithBackend(backend))
	}

	return nil
}

// BuildWithBackend - returns new HTTPFrontend cache with given backend
func (f *FrontendFactory) BuildWithBackend(backend Backend) *Frontend {
	frontend := f.provider()
	frontend.backend = backend

	return frontend
}

// BuildBackend by given BackendConfig and frontendName
//
//nolint:cyclop // it is what it is
func (f *FrontendFactory) BuildBackend(backendConfig BackendConfig, frontendName string) (Backend, error) {
	switch backendConfig.BackendType {
	case "redis":
		if backendConfig.Redis == nil {
			return nil, ErrRedisConfig
		}

		return f.NewRedisBackend(*backendConfig.Redis, frontendName)
	case "memory":
		if backendConfig.Memory == nil {
			return nil, ErrMemoryConfig
		}

		return f.NewMemoryBackend(*backendConfig.Memory, frontendName)
	case "twolevel":
		if backendConfig.Twolevel == nil || backendConfig.Twolevel.First == nil || backendConfig.Twolevel.Second == nil {
			return nil, ErrTwoLevelConfig
		}

		first, err := f.BuildBackend(*backendConfig.Twolevel.First, frontendName)
		if err != nil {
			return nil, err
		}

		second, err := f.BuildBackend(*backendConfig.Twolevel.Second, frontendName)
		if err != nil {
			return nil, err
		}

		return f.NewTwoLevel(TwoLevelBackendConfig{first, second})
	}

	return nil, fmt.Errorf("backend type %q error: %w", backendConfig.BackendType, ErrInvalidBackend)
}

// NewMemoryBackend with given config and name
func (f *FrontendFactory) NewMemoryBackend(config MemoryBackendConfig, frontendName string) (Backend, error) {
	return f.inMemoryBackendFactory.SetConfig(config).SetFrontendName(frontendName).Build()
}

// NewRedisBackend with given config and name
func (f *FrontendFactory) NewRedisBackend(config RedisBackendConfig, frontendName string) (Backend, error) {
	return f.redisBackendFactory.SetConfig(config).SetFrontendName(frontendName).Build()
}

// NewTwoLevel with given config
func (f *FrontendFactory) NewTwoLevel(config TwoLevelBackendConfig) (Backend, error) {
	return f.twoLevelBackendFactory.SetConfig(config).Build()
}
