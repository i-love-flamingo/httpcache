package httpcache

import (
	"context"

	"flamingo.me/flamingo/v3/framework/opencensus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	backendTypeCacheKeyType, _    = tag.NewKey("backend_type")
	frontendNameCacheKeyType, _   = tag.NewKey("frontend_name")
	backendCacheKeyErrorReason, _ = tag.NewKey("error_reason")
	backendCacheHitCount          = stats.Int64("flamingo/httpcache/backend/hit", "Count of cache-backend hits", stats.UnitDimensionless)
	backendCacheMissCount         = stats.Int64("flamingo/httpcache/backend/miss", "Count of cache-backend misses", stats.UnitDimensionless)
	backendCacheErrorCount        = stats.Int64("flamingo/httpcache/backend/error", "Count of cache-backend errors", stats.UnitDimensionless)
	backendCacheEntriesCount      = stats.Int64("flamingo/httpcache/backend/entries", "Count of cache-backend entries", stats.UnitDimensionless)
)

type (

	// MetricsBackend - a Backend that logs metrics
	MetricsBackend struct {
	}

	// Metrics take care of publishing metrics for a specific cache
	Metrics struct {
		// backendType - the type of the cache backend
		backendType string
		// frontendName - the name if the cache frontend where the backend is attached
		frontendName string
	}
)

// NewCacheMetrics creates a backend metrics helper instance
func NewCacheMetrics(backendType string, frontendName string) Metrics {
	b := Metrics{
		backendType:  backendType,
		frontendName: frontendName,
	}

	return b
}

func init() {
	if err := opencensus.View(
		"flamingo/httpcache/backend/hit",
		backendCacheHitCount,
		view.Count(),
		backendTypeCacheKeyType,
		frontendNameCacheKeyType,
	); err != nil {
		panic(err)
	}

	if err := opencensus.View(
		"flamingo/httpcache/backend/miss",
		backendCacheMissCount,
		view.Count(),
		backendTypeCacheKeyType,
		frontendNameCacheKeyType,
	); err != nil {
		panic(err)
	}

	if err := opencensus.View(
		"flamingo/httpcache/backend/error",
		backendCacheErrorCount,
		view.Count(),
		backendTypeCacheKeyType,
		frontendNameCacheKeyType,
		backendCacheKeyErrorReason,
	); err != nil {
		panic(err)
	}

	if err := opencensus.View(
		"flamingo/httpcache/backend/entries",
		backendCacheEntriesCount,
		view.LastValue(),
		backendTypeCacheKeyType,
		frontendNameCacheKeyType,
	); err != nil {
		panic(err)
	}
}

func (bi Metrics) countHit() {
	ctx, _ := tag.New(
		context.Background(),
		tag.Upsert(opencensus.KeyArea, "cacheBackend"),
		tag.Upsert(backendTypeCacheKeyType, bi.backendType),
		tag.Upsert(frontendNameCacheKeyType, bi.frontendName),
	)
	stats.Record(ctx, backendCacheHitCount.M(1))
}

func (bi Metrics) countMiss() {
	ctx, _ := tag.New(
		context.Background(),
		tag.Upsert(opencensus.KeyArea, "cacheBackend"),
		tag.Upsert(backendTypeCacheKeyType, bi.backendType),
		tag.Upsert(frontendNameCacheKeyType, bi.frontendName),
	)
	stats.Record(ctx, backendCacheMissCount.M(1))
}

func (bi Metrics) countError(reason string) {
	ctx, _ := tag.New(
		context.Background(),
		tag.Upsert(opencensus.KeyArea, "cacheBackend"),
		tag.Upsert(backendTypeCacheKeyType, bi.backendType),
		tag.Upsert(frontendNameCacheKeyType, bi.frontendName),
		tag.Upsert(backendCacheKeyErrorReason, reason),
	)
	stats.Record(ctx, backendCacheErrorCount.M(1))
}

func (bi Metrics) recordEntries(entries int64) {
	ctx, _ := tag.New(
		context.Background(),
		tag.Upsert(opencensus.KeyArea, "cacheBackend"),
		tag.Upsert(backendTypeCacheKeyType, bi.backendType),
		tag.Upsert(frontendNameCacheKeyType, bi.frontendName),
	)
	stats.Record(ctx, backendCacheEntriesCount.M(entries))
}
