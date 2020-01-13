package httpcache

import (
	"context"
	"time"

	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/golang/groupcache/singleflight"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

type (
	// Frontend caches and delivers HTTP responses
	Frontend struct {
		singleflight.Group
		backend Backend
		logger  flamingo.Logger
	}
)

// Inject dependencies
func (f *Frontend) Inject(
	logger flamingo.Logger,
) *Frontend {
	f.logger = logger.WithField(flamingo.LogKeyCategory, "httpcache")

	return f
}

// SetBackend for usage
func (f *Frontend) SetBackend(b Backend) {
	f.backend = b
}

// Get the cached response if possible or perform a call to loader
// The result of loader will be returned and cached
func (f *Frontend) Get(ctx context.Context, key string, loader HTTPLoader) (Entry, error) {
	if f.backend == nil {
		return Entry{}, errors.New("no backend defined")
	}

	ctx, span := trace.StartSpan(ctx, "flamingo/httpcache/get")
	span.Annotate(nil, key)
	defer span.End()

	if entry, ok := f.backend.Get(key); ok {
		if entry.Meta.LifeTime.After(time.Now()) {
			f.logger.Debug("Serving from cache: ", key)
			return entry, nil
		}

		if entry.Meta.GraceTime.After(time.Now()) {
			// Try to load the actual value in background
			go func() {
				_, _ = f.load(context.Background(), key, loader)
			}()
			f.logger.Debug("Gracetime! Serving from cache: ", key)
			return entry, nil
		}
	}
	f.logger.Debug("no cache entry for: ", key)

	return f.load(ctx, key, loader)
}

func (f *Frontend) load(ctx context.Context, key string, loader HTTPLoader) (Entry, error) {
	ctx, span := trace.StartSpan(ctx, "flamingo/httpcache/load")
	span.Annotate(nil, key)
	defer span.End()

	data, err := f.Do(key, func() (res interface{}, resultErr error) {
		ctx, fetchRoutineSpan := trace.StartSpan(context.Background(), "flamingo/httpcache/fetchRoutine")
		fetchRoutineSpan.Annotate(nil, key)
		defer fetchRoutineSpan.End()

		defer func() {
			if err := recover(); err != nil {
				if err2, ok := err.(error); ok {
					resultErr = errors.WithStack(err2)
				} else {
					resultErr = errors.WithStack(errors.Errorf("HTTPCache Frontend.load exception: %#v", err))
				}
			}
		}()

		entry, err := loader(ctx)
		if err != nil {
			return nil, err
		}

		return entry, err
	})
	if err != nil {
		return Entry{}, err
	}
	entry := data.(Entry)

	f.logger.WithContext(ctx).Debugf("Store entry in Cache for key: %s", key)
	_ = f.backend.Set(key, entry)

	return entry, err
}
