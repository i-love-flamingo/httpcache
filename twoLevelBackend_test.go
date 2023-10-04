package httpcache_test

import (
	"testing"
	"time"

	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/stretchr/testify/assert"

	"flamingo.me/httpcache"
)

func createInMemoryBackend() httpcache.Backend {
	return func() httpcache.Backend {
		f := httpcache.InMemoryBackendFactory{}
		backend, _ := f.SetConfig(httpcache.MemoryBackendConfig{Size: 100}).SetFrontendName("default").SetLurkerPeriod(100 * time.Millisecond).Build()

		return backend
	}()
}

func Test_RunDefaultBackendTestCase_TwoLevelBackend(t *testing.T) {
	t.Parallel()

	levelBackendFactory := httpcache.TwoLevelBackendFactory{}
	c := httpcache.TwoLevelBackendConfig{
		FirstLevel:  createInMemoryBackend(),
		SecondLevel: createInMemoryBackend(),
	}

	backend, err := levelBackendFactory.Inject(flamingo.NullLogger{}).SetConfig(c).Build()
	assert.NoError(t, err)
	testcase := NewBackendTestCase(t, backend, true)
	testcase.RunTests()
}
