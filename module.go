package httpcache

import (
	"flamingo.me/dingo"
	"flamingo.me/flamingo/v3/framework/config"
)

type (
	// Module basic struct
	Module struct {
		frontendFactory *FrontendFactory
		cacheConfig     config.Map
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
	return `
// httpcache config
httpcache: {
	Redis :: {
		backendType: "redis"
		redis: {
			host:               string | *"localhost"
			port:               string | *"6379"
			idleTimeOutSeconds: int | *60
			maxIdle:            int | *8
		}
	}

	Memory :: {
		backendType: "memory"
		memory:
			size: int | *200
	}

	Twolevel :: {
		backendType: "twolevel"
		twolevel: {
			first:  Cache
			second: Cache
		}
	}

	Cache :: {backendType: string} & (Redis | Memory | Twolevel)

	frontendFactory: {
		[string]: Cache
	}
}
`
}
