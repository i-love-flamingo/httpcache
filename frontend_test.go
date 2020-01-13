package httpcache_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"flamingo.me/flamingo/v3/framework/flamingo"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"flamingo.me/httpcache"
	"flamingo.me/httpcache/mocks"
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wait := make(chan struct{}, 1)
			loaderCalled := false
			loader := func(ctx context.Context) (httpcache.Entry, error) {
				loaderCalled = true

				return tt.args.loader(ctx)
			}

			backend := new(mocks.Backend)
			if tt.cacheGetter != nil {
				backend.On("Get", testKey).Return(tt.cacheGetter())
			}
			if tt.wantSet != nil {
				backend.On("Set", testKey, *tt.wantSet).Run(func(_ mock.Arguments) {
					wait <- struct{}{}
				}).Return(nil).Once()
			} else {
				close(wait)
			}

			f := new(httpcache.Frontend).Inject(new(flamingo.NullLogger))
			f.SetBackend(backend)
			got, err := f.Get(context.Background(), testKey, loader)

			// wait for eventually async cache set to be finished
			<-wait

			assert.Equal(t, tt.wantLoaderCalled, loaderCalled, "Loader expected to be called: %v, but actually called: %v", tt.wantLoaderCalled, loaderCalled)

			backend.AssertExpectations(t)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Error("expected entry is wrong: ", diff)
			}
		})
	}
}
