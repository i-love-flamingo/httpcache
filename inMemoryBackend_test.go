package httpcache_test

import (
	"testing"

	"flamingo.me/httpcache"
)

func Test_RunDefaultBackendTestCase_InMemoryBackend(t *testing.T) {
	f := httpcache.InMemoryBackendFactory{}
	backend, _ := f.SetConfig(httpcache.MemoryBackendConfig{Size: 100}).SetFrontendName("default").Build()

	testCase := NewBackendTestCase(t, backend, true)
	testCase.RunTests()
}
