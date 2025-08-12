//go:build integration

package httpcache_test

import (
	"testing"

	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/stretchr/testify/assert"

	"flamingo.me/httpcache"
)

func TestHTTPFrontendFactory_BuildBackend_Docker(t *testing.T) {
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

	t.Run("redis", func(t *testing.T) {
		t.Parallel()

		testConfig := httpcache.BackendConfig{
			BackendType: "redis",
			Redis: &httpcache.RedisBackendConfig{
				IdleTimeOutSeconds: 1,
				Host:               redisHost,
				Port:               redisPort,
				Username:           username,
				Password:           password,
			},
		}

		backend, err := factory.BuildBackend(testConfig, "test")
		assert.NoError(t, err)
		assert.IsType(t, &httpcache.RedisBackend{}, backend)
	})
}
