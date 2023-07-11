package httpcache

import (
	"context"
	"time"
)

type (
	// Backend to persist cache data
	Backend interface {
		Get(key string) (Entry, bool)
		Set(key string, entry Entry) error
		Purge(key string) error
		Flush() error
	}

	// TagSupporting describes a cache backend, responsible for storing, flushing, setting and getting entries
	TagSupporting interface {
		PurgeTags(tags []string) error
	}

	// Entry represents a cached HTTP Response
	Entry struct {
		Meta       Meta
		Header     map[string][]string
		Status     string
		StatusCode int
		Body       []byte
	}

	// Meta data for a cache Entry
	Meta struct {
		LifeTime  time.Time
		GraceTime time.Time
		Tags      []string
	}

	// HTTPLoader returns an Entry to be cached. All Entries will be cached if error is nil
	HTTPLoader func(context.Context) (Entry, error)
)
