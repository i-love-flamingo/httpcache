package httpcache_test

import (
	"testing"

	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3/framework/config"
	"flamingo.me/flamingo/v3/framework/flamingo"

	"flamingo.me/httpcache"
)

type (
	loggerModule struct{}
)

// Configure DI
func (l *loggerModule) Configure(injector *dingo.Injector) {
	injector.Bind(new(flamingo.Logger)).To(new(flamingo.NullLogger))
}

func TestModule_Configure(t *testing.T) {
	if err := config.TryModules(nil, new(loggerModule), new(httpcache.Module)); err != nil {
		t.Error(err)
	}
}
