package httpcache_test

import (
	"testing"
	"time"

	"flamingo.me/httpcache"
)

func Test_RunDefaultBackendTestCase_InMemoryBackend(t *testing.T) {
	t.Parallel()

	f := httpcache.InMemoryBackendFactory{}
	backend, _ := f.SetConfig(httpcache.MemoryBackendConfig{Size: 100}).SetFrontendName("default").SetLurkerPeriod(100 * time.Millisecond).Build()

	testCase := NewBackendTestCase(t, backend, true)
	testCase.RunTests()
}
