package httpcache

import (
	"context"
	"fmt"
	"time"

	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"golang.org/x/sync/singleflight"
)

var ErrInvalidEntry = errors.New("cache returned invalid entry type")

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
	f.logger = logger

	return f
}

// SetBackend for usage
func (f *Frontend) SetBackend(b Backend) *Frontend {
	f.backend = b

	return f
}

func (f *Frontend) Purge(ctx context.Context, key string) error {
	if f.backend == nil {
		return errors.New("no backend defined")
	}

	_, span := trace.StartSpan(ctx, "flamingo/httpcache/purge")

	span.Annotate(nil, key)

	err := f.backend.Purge(key)
	if err != nil {
		return fmt.Errorf("failed to purge with key: %s: %w", key, err)
	}

	return nil
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
			f.logger.WithContext(ctx).
				WithField(flamingo.LogKeyCategory, "httpcache").
				Debug("Serving from cache: ", key)

			return entry, nil
		}

		if entry.Meta.GraceTime.After(time.Now()) {
			// Try to load the actual value in background
			go func() {
				_, _ = f.load(ctx, key, loader)
			}()

			f.logger.WithContext(ctx).
				WithField(flamingo.LogKeyCategory, "httpcache").
				Debug("Gracetime! Serving from cache: ", key)

			return entry, nil
		}
	}

	f.logger.WithContext(ctx).
		WithField(flamingo.LogKeyCategory, "httpcache").
		Debug("No cache entry for: ", key)

	return f.load(ctx, key, loader)
}

func (f *Frontend) load(ctx context.Context, key string, loader HTTPLoader) (Entry, error) {
	oldSpan := trace.FromContext(ctx)
	newContext := trace.NewContext(context.Background(), oldSpan)

	newContextWithSpan, span := trace.StartSpan(newContext, "flamingo/httpcache/load")

	span.Annotate(nil, key)

	defer span.End()

	data, err, _ := f.Do(key, func() (res interface{}, resultErr error) {
		ctx, fetchRoutineSpan := trace.StartSpan(
			newContextWithSpan,
			"flamingo/httpcache/fetchRoutine",
		)
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

		f.logger.WithContext(newContextWithSpan).
			WithField(flamingo.LogKeyCategory, "httpcache").
			Debugf("Store entry in Cache for key: %s", key)

		_ = f.backend.Set(key, entry)

		return entry, err
	})
	if err != nil {
		return Entry{}, fmt.Errorf("http loader error: %w", err)
	}

	entry, ok := data.(Entry)
	if !ok {
		return Entry{}, ErrInvalidEntry
	}

	return entry, nil
}
