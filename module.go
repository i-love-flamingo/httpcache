package httpcache

import (
	"flamingo.me/dingo"
)

type (
	// Module basic struct
	Module struct {
		frontendFactory *FrontendFactory
	}
)

// Inject dependencies
func (m *Module) Inject(
	frontendFactory *FrontendFactory,
) *Module {
	m.frontendFactory = frontendFactory

	return m
}

// Configure DI
func (m *Module) Configure(injector *dingo.Injector) {
	err := m.frontendFactory.BindConfiguredCaches(injector)
	if err != nil {
		panic(err)
	}
}

// CueConfig definition
func (m *Module) CueConfig() string {
	// language=cue
	return `
// httpcache config
httpcache: {
	Redis :: {
		backendType: "redis"
		redis: {
			host:               string | *"localhost"
			port:               string | *"6379"
			idleTimeOutSeconds: int | float | *60
			maxIdle:            int | float | *8
		}
	}

	Memory :: {
		backendType: "memory"
		memory: {
			size: int | float | *200
		}
	}

	Twolevel :: {
		backendType: "twolevel"
		twolevel: {
			first:  Cache
			second: Cache
		}
	}

	Cache :: Redis | Memory | Twolevel

	frontendFactory: {
		[string]: Cache
	}
}
`
}
