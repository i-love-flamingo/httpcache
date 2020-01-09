package httpcache_test

import (
	"encoding/gob"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"flamingo.me/httpcache"
)

type backendTestEntry struct {
	Content string
}

type (
	// BackendTestCase representations
	BackendTestCase struct {
		t            *testing.T
		backend      httpcache.Backend
		tagsInResult bool
	}
)

func init() {
	gob.Register(new(backendTestEntry))
}

func NewBackendTestCase(t *testing.T, backend httpcache.Backend, tagsInResult bool) *BackendTestCase {
	return &BackendTestCase{
		t:            t,
		backend:      backend,
		tagsInResult: tagsInResult,
	}
}

func (tc *BackendTestCase) RunTests() {
	tc.testSetGetPurge()

	tc.testFlush()

	if _, ok := tc.backend.(httpcache.TagSupporting); ok {
		tc.testPurgeTags()
	}
}

func (tc *BackendTestCase) testSetGetPurge() {
	entry := tc.buildEntry("ASDF", []string{"eins", "zwei"})
	wantedEntry := entry

	tc.setAndCompareEntry("ONE_KEY", entry, wantedEntry)
	tc.setAndCompareEntry("ANOTHER_KEY", entry, wantedEntry)

	err := tc.backend.Purge("ONE_KEY")
	if err != nil {
		tc.t.Fatalf("Purge Key Failed: %v", err)
	}

	tc.shouldNotExist("ONE_KEY")

	tc.getAndCompareEntry("ANOTHER_KEY", wantedEntry)
}

func (tc *BackendTestCase) testFlush() {
	entry := tc.buildEntry("ASDF", []string{"eins", "zwei"})

	tc.setEntry("ONE_KEY", entry)
	tc.setEntry("ANOTHERKEY_KEY", entry)

	err := tc.backend.Flush()
	if err != nil {
		tc.t.Fatalf("Flush Failed: %v", err)
	}

	tc.shouldNotExist("ONE_KEY")
	tc.shouldNotExist("ANOTHERKEY_KEY")
}

func (tc *BackendTestCase) testPurgeTags() {
	entry := tc.buildEntry("ASDF", []string{"eins", "zwei"})
	entryWithoutTags := tc.buildEntry("ASDF", []string{})

	tc.setEntry("ONE_KEY", entry)
	tc.setEntry("ANOTHERKEY_KEY", entry)
	tc.setEntry("THIRD_KEY", entryWithoutTags)

	tagsToPurge := []string{"eins"}
	err := tc.backend.(httpcache.TagSupporting).PurgeTags(tagsToPurge)
	if err != nil {
		tc.t.Fatalf("Purge Tags Failed: %v", err)
	}

	tc.shouldNotExist("ONE_KEY")
	tc.shouldNotExist("ANOTHERKEY_KEY")
	tc.shouldExist("THIRD_KEY")
}

func (tc *BackendTestCase) setEntry(key string, entry httpcache.Entry) {
	err := tc.backend.Set(key, entry)
	if err != nil {
		tc.t.Fatalf("Failed to set Entry for key %v with error: %v", key, err)
	}
	tc.shouldExist(key)
}

func (tc *BackendTestCase) setAndCompareEntry(key string, entry httpcache.Entry, wanted httpcache.Entry) {
	tc.setEntry(key, entry)
	tc.getAndCompareEntry(key, wanted)
}

func (tc *BackendTestCase) getAndCompareEntry(key string, wanted httpcache.Entry) {
	entry := tc.shouldExist(key)
	tc.mustBeEqual(entry, wanted)
}

func (tc *BackendTestCase) mustBeEqual(entry httpcache.Entry, wanted httpcache.Entry) {
	require.True(tc.t, entry.Meta.GraceTime.Equal(wanted.Meta.GraceTime), "Entry gracetimes are not equal. got: %v, want %v", entry.Meta.GraceTime, wanted.Meta.GraceTime)
	require.True(tc.t, entry.Meta.LifeTime.Equal(wanted.Meta.LifeTime), "Entry lifetimes are not equal. got: %v, want %v", entry.Meta.LifeTime, wanted.Meta.LifeTime)
	require.Equal(tc.t, entry.Meta.Tags, wanted.Meta.Tags, "Entry Meta.Tags are not equal. got: %v, want %v", entry.Meta.Tags, wanted.Meta.Tags)
	require.Equal(tc.t, entry.Body, wanted.Body, "Entry body are not equal. got: %s, want %s", entry.Body, wanted.Body)
	require.Equal(tc.t, entry.Status, wanted.Status, "Entry status are not equal. got: %s, want %s", entry.Status, wanted.Status)
	require.Equal(tc.t, entry.StatusCode, wanted.StatusCode, "Entry status code are not equal. got: %v, want %v", entry.StatusCode, wanted.StatusCode)
}

func (tc *BackendTestCase) shouldExist(key string) httpcache.Entry {
	entry, found := tc.backend.Get(key)
	if !found {
		tc.t.Fatalf("Failed to get Entry with key: %v", key)
	}
	return entry
}

func (tc *BackendTestCase) shouldNotExist(key string) {
	entry, found := tc.backend.Get(key)
	if found {
		tc.t.Fatalf("Entry with key %v should not exists, but returns %v", key, entry)
	}
}

func (tc *BackendTestCase) buildEntry(content string, tags []string) httpcache.Entry {
	return httpcache.Entry{
		Meta: httpcache.Meta{
			LifeTime:  time.Now().Add(time.Minute * 3),
			GraceTime: time.Now().Add(time.Minute * 30),
			Tags:      tags,
		},
		Header:     nil,
		Status:     "",
		StatusCode: 0,
		Body:       []byte(content),
	}
}
