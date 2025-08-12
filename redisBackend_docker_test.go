//go:build integration

package httpcache_test

import (
	"testing"

	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/stretchr/testify/assert"

	"flamingo.me/httpcache"
)

func Test_RunDefaultBackendTestCase_RedisBackend(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		config := httpcache.RedisBackendConfig{
			MaxIdle:            8,
			IdleTimeOutSeconds: 30,
			Host:               redisHost,
			Port:               redisPort,
			Username:           username,
			Password:           password,
		}

		factory := httpcache.RedisBackendFactory{}

		backend, err := factory.Inject(flamingo.NullLogger{}).SetConfig(config).SetFrontendName("testfrontend").Build()
		assert.NoError(t, err)
		testcase := NewBackendTestCase(t, backend, false)
		testcase.RunTests()
	})

	t.Run("failure", func(t *testing.T) {
		config := httpcache.RedisBackendConfig{
			MaxIdle:            8,
			IdleTimeOutSeconds: 30,
			Host:               redisHost,
			Port:               redisPort,
			Username:           "username",
			Password:           "password",
		}

		factory := httpcache.RedisBackendFactory{}

		_, err := factory.Inject(flamingo.NullLogger{}).SetConfig(config).SetFrontendName("testfrontend").Build()
		assert.Error(t, err)
	})
}
