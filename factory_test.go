package httpcache_test

import (
	"testing"

	"flamingo.me/flamingo/v3/framework/config"
	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"flamingo.me/httpcache"
)

func TestHTTPFrontendFactory_ConfigUnmarshalling(t *testing.T) {
	t.Parallel()

	testconfig := config.Map{
		"one": config.Map{
			"backendType": "inmemory",
			"Memory": config.Map{
				"size": 100.0,
			},
		},
		"two": config.Map{
			"backendType": "inmemory",
			"Memory": config.Map{
				"size": 100.0,
			},
		},
	}

	var typedCacheConfig httpcache.FactoryConfig

	require.NoError(t, testconfig.MapInto(&typedCacheConfig))

	assert.Contains(t, typedCacheConfig, "one")
	assert.Contains(t, typedCacheConfig, "two")

	one := typedCacheConfig["one"]
	assert.Equal(t, "inmemory", one.BackendType)
	require.NotNil(t, one.Memory)
	assert.Equal(t, one.Memory.Size, 100)
}

func TestHTTPFrontendFactory_BuildBackend(t *testing.T) {
	t.Parallel()

	provider := func() *httpcache.Frontend {
		return new(httpcache.Frontend)
	}

	factory := &httpcache.FrontendFactory{}
	factory.Inject(
		provider,
		new(httpcache.RedisBackendFactory).Inject(new(flamingo.NullLogger)),
		&httpcache.InMemoryBackendFactory{},
		&httpcache.TwoLevelBackendFactory{},
		nil,
	)

	t.Run("memory", func(t *testing.T) {
		t.Parallel()

		testConfig := httpcache.BackendConfig{
			BackendType: "memory",
			Memory:      &httpcache.MemoryBackendConfig{Size: 10},
		}

		backend, err := factory.BuildBackend(testConfig, "test")
		assert.NoError(t, err)
		assert.IsType(t, &httpcache.MemoryBackend{}, backend)
	})

	t.Run("inmemory error", func(t *testing.T) {
		t.Parallel()

		testConfig := httpcache.BackendConfig{
			BackendType: "memory",
		}

		_, err := factory.BuildBackend(testConfig, "test")
		assert.Error(t, err)
	})
}
