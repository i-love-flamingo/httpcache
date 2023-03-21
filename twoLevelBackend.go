package httpcache

import (
	"errors"
	"fmt"

	"flamingo.me/flamingo/v3/framework/flamingo"
)

var ErrAllBackendsFailed = errors.New("all backends failed")
var ErrAtLeastOneBackendFailed = errors.New("at least one backends failed")

type (
	// TwoLevelBackend the cache backend interface with a two level solution
	TwoLevelBackend struct {
		firstBackend  Backend
		secondBackend Backend
		logger        flamingo.Logger
	}

	// TwoLevelBackendConfig defines the backends to be used
	TwoLevelBackendConfig struct {
		FirstLevel  Backend
		SecondLevel Backend
	}

	// TwoLevelBackendFactory creates instances of TwoLevel backends
	TwoLevelBackendFactory struct {
		logger flamingo.Logger
		config TwoLevelBackendConfig
	}
)

var _ Backend = new(TwoLevelBackend)

// Inject dependencies
func (f *TwoLevelBackendFactory) Inject(logger flamingo.Logger) *TwoLevelBackendFactory {
	f.logger = logger
	return f
}

// SetConfig for factory
func (f *TwoLevelBackendFactory) SetConfig(config TwoLevelBackendConfig) *TwoLevelBackendFactory {
	f.config = config
	return f
}

// Build the instance
func (f *TwoLevelBackendFactory) Build() (Backend, error) {
	return &TwoLevelBackend{
		firstBackend:  f.config.FirstLevel,
		secondBackend: f.config.SecondLevel,
		logger:        f.logger,
	}, nil
}

// Get entry by key
func (mb *TwoLevelBackend) Get(key string) (entry Entry, found bool) {
	entry, found = mb.firstBackend.Get(key)
	if found {
		return entry, found
	}

	entry, found = mb.secondBackend.Get(key)
	if found {
		go func() {
			_ = mb.firstBackend.Set(key, entry)
		}()

		return entry, true
	}

	return Entry{}, false
}

// Set entry for key
func (mb *TwoLevelBackend) Set(key string, entry Entry) error {
	errorCount := 0

	err := mb.firstBackend.Set(key, entry)
	if err != nil {
		errorCount++

		mb.logger.WithField("category", "TwoLevelBackend").Error(fmt.Sprintf("Failed to set key %v with error %v", key, err))
	}

	err = mb.secondBackend.Set(key, entry)
	if err != nil {
		errorCount++

		mb.logger.WithField("category", "TwoLevelBackend").Error(fmt.Sprintf("Failed to set key %v with error %v", key, err))
	}

	if errorCount >= 2 { //nolint:gomnd // there are two backends no need to introduce const for that
		return ErrAllBackendsFailed
	}

	return nil
}

// Purge entry by key
func (mb *TwoLevelBackend) Purge(key string) (err error) {
	var errorList []error

	err = mb.firstBackend.Purge(key)
	if err != nil {
		errorList = append(errorList, err)
		mb.logger.WithField("category", "TwoLevelBackend").Error(fmt.Sprintf("Failed Purge with error %v", err))
	}

	err = mb.secondBackend.Purge(key)
	if err != nil {
		errorList = append(errorList, err)
		mb.logger.WithField("category", "TwoLevelBackend").Error(fmt.Sprintf("Failed Purge with error %v", err))
	}

	if 0 != len(errorList) {
		return fmt.Errorf("not all backends succeeded to Purge key %v, errors: %v - %w", key, errorList, ErrAtLeastOneBackendFailed)
	}

	return nil
}

// Flush the whole cache
func (mb *TwoLevelBackend) Flush() (err error) {
	var errorList []error

	err = mb.firstBackend.Flush()
	if err != nil {
		errorList = append(errorList, err)
		mb.logger.WithField("category", "TwoLevelBackend").Error(fmt.Sprintf("Failed Flush error %v", err))
	}

	err = mb.secondBackend.Flush()
	if err != nil {
		errorList = append(errorList, err)
		mb.logger.WithField("category", "TwoLevelBackend").Error(fmt.Sprintf("Failed Flush error %v", err))
	}

	if 0 != len(errorList) {
		return fmt.Errorf("not all backends succeeded to Flush. errors: %v - %w", errorList, ErrAtLeastOneBackendFailed)
	}

	return nil
}
