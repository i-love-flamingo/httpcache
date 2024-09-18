package httpcache_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"flamingo.me/httpcache/mocks"

	"flamingo.me/httpcache"
)

var (
	testKey = "testkey"
)

func createEntry(t *testing.T, lifeTime, graceTime string, tags []string, header map[string][]string, status string, statusCode int, body string) httpcache.Entry {
	t.Helper()

	lifeTimeDuration, err := time.ParseDuration(lifeTime)
	require.NoError(t, err)
	graceTimeDuration, err := time.ParseDuration(graceTime)
	require.NoError(t, err)

	return httpcache.Entry{
		Meta: httpcache.Meta{
			LifeTime:  time.Now().Add(lifeTimeDuration),
			GraceTime: time.Now().Add(graceTimeDuration),
			Tags:      tags,
		},
		Header:     header,
		Status:     status,
		StatusCode: statusCode,
		Body:       []byte(body),
	}
}

func TestFrontend_Get(t *testing.T) {
	t.Parallel()
	defaultEntry := createEntry(t, "10s", "15s", nil, nil, "200 OK", http.StatusOK, "default")
	defaultGraceEntry := createEntry(t, "-10s", "15s", nil, nil, "200 OK", http.StatusOK, "grace")
	defaultOldEntry := createEntry(t, "-10s", "-15s", nil, nil, "200 OK", http.StatusOK, "old")

	type args struct {
		loader httpcache.HTTPLoader
	}

	tests := []struct {
		name             string
		cacheGetter      func() (httpcache.Entry, bool)
		args             args
		want             httpcache.Entry
		wantSet          *httpcache.Entry
		wantErr          bool
		wantLoaderCalled bool
	}{
		{
			name:             "in cache and in lifetime",
			cacheGetter:      func() (httpcache.Entry, bool) { return defaultEntry, true },
			want:             defaultEntry,
			wantErr:          false,
			wantLoaderCalled: false,
		},
		{
			name:        "not in cache",
			cacheGetter: func() (httpcache.Entry, bool) { return httpcache.Entry{}, false },
			args: args{
				loader: func(_ context.Context) (httpcache.Entry, error) {
					return defaultEntry, nil
				}},
			want:             defaultEntry,
			wantSet:          &defaultEntry,
			wantErr:          false,
			wantLoaderCalled: true,
		},
		{
			name:        "in cache but not in lifetime/gracetime",
			cacheGetter: func() (httpcache.Entry, bool) { return defaultOldEntry, true },
			args: args{
				loader: func(_ context.Context) (httpcache.Entry, error) {
					return defaultEntry, nil
				}},
			want:             defaultEntry,
			wantSet:          &defaultEntry,
			wantErr:          false,
			wantLoaderCalled: true,
		},
		{
			name:        "in cache, not in lifetime, in gracetime",
			cacheGetter: func() (httpcache.Entry, bool) { return defaultGraceEntry, true },
			args: args{
				loader: func(_ context.Context) (httpcache.Entry, error) {
					return defaultEntry, nil
				}},
			want:             defaultGraceEntry,
			wantSet:          &defaultEntry,
			wantErr:          false,
			wantLoaderCalled: true,
		},
		{
			name:        "not in cache, loader panics",
			cacheGetter: func() (httpcache.Entry, bool) { return httpcache.Entry{}, false },
			args: args{
				loader: func(_ context.Context) (httpcache.Entry, error) {
					panic("test panic")
				}},
			want:             httpcache.Entry{},
			wantErr:          true,
			wantLoaderCalled: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			wait := make(chan struct{}, 1)
			loaderCalled := false
			loader := func(ctx context.Context) (httpcache.Entry, error) {
				loaderCalled = true

				return test.args.loader(ctx)
			}

			backend := new(mocks.Backend)

			if test.cacheGetter != nil {
				backend.EXPECT().Get(testKey).Return(test.cacheGetter())
			}

			if test.wantSet != nil {
				backend.EXPECT().Set(testKey, *test.wantSet).Run(func(key string, entry httpcache.Entry) {
					wait <- struct{}{}
				}).Return(nil).Once()
			} else {
				close(wait)
			}

			f := new(httpcache.Frontend).Inject(new(flamingo.NullLogger)).SetBackend(backend)
			got, err := f.Get(context.Background(), testKey, loader)

			// wait for eventually async cache set to be finished
			<-wait

			assert.Equal(t, test.wantLoaderCalled, loaderCalled, "Loader expected to be called: %v, but actually called: %v", test.wantLoaderCalled, loaderCalled)

			backend.AssertExpectations(t)

			if (err != nil) != test.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, test.wantErr)

				return
			}

			assert.Equal(t, test.want, got)
		})
	}
}

func TestContextDeadlineExceeded(t *testing.T) {
	t.Parallel()

	t.Run("exceeded, throw error", func(t *testing.T) {
		t.Parallel()

		backend := new(mocks.Backend)

		backend.EXPECT().Get(testKey).Return(func() (httpcache.Entry, bool) { return httpcache.Entry{}, false }())

		backend.EXPECT().Set(mock.Anything, mock.Anything).Return(nil)

		contextWithDeadline, cancel := context.WithDeadline(context.Background(), time.Now().Add(4*time.Second))
		t.Cleanup(cancel)

		f := new(httpcache.Frontend).Inject(new(flamingo.NullLogger)).SetBackend(backend)
		got, err := f.Get(contextWithDeadline, testKey, loaderWithWaitingTime)

		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.Equal(t, httpcache.Entry{}, got)
	})

	t.Run("did not exceed, no error", func(t *testing.T) {
		t.Parallel()

		backend := new(mocks.Backend)

		backend.EXPECT().Get(testKey).Return(func() (httpcache.Entry, bool) { return httpcache.Entry{}, false }())

		backend.EXPECT().Set(mock.Anything, mock.Anything).Return(nil)

		contextWithDeadline, cancel := context.WithDeadline(context.Background(), time.Now().Add(6*time.Second))
		t.Cleanup(cancel)

		f := new(httpcache.Frontend).Inject(new(flamingo.NullLogger)).SetBackend(backend)
		got, err := f.Get(contextWithDeadline, testKey, loaderWithWaitingTime)

		assert.NoError(t, err)
		assert.Equal(t, []byte("body"), got.Body)
	})
}

func loaderWithWaitingTime(ctx context.Context) (httpcache.Entry, error) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)

		_, _ = w.Write([]byte("Test 123"))
	}))

	defer server.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		return httpcache.Entry{}, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if resp != nil {
		_ = resp.Body.Close()
	}

	if err != nil {
		return httpcache.Entry{}, err
	}

	return httpcache.Entry{
		Meta: httpcache.Meta{
			LifeTime:  time.Now().Add(10),
			GraceTime: time.Now().Add(15),
			Tags:      nil,
		},
		Header:     nil,
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Body:       []byte("body"),
	}, nil
}
