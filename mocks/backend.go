// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	httpcache "flamingo.me/httpcache"
	mock "github.com/stretchr/testify/mock"
)

// Backend is an autogenerated mock type for the Backend type
type Backend struct {
	mock.Mock
}

type Backend_Expecter struct {
	mock *mock.Mock
}

func (_m *Backend) EXPECT() *Backend_Expecter {
	return &Backend_Expecter{mock: &_m.Mock}
}

// Flush provides a mock function with no fields
func (_m *Backend) Flush() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Flush")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Backend_Flush_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Flush'
type Backend_Flush_Call struct {
	*mock.Call
}

// Flush is a helper method to define mock.On call
func (_e *Backend_Expecter) Flush() *Backend_Flush_Call {
	return &Backend_Flush_Call{Call: _e.mock.On("Flush")}
}

func (_c *Backend_Flush_Call) Run(run func()) *Backend_Flush_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Backend_Flush_Call) Return(_a0 error) *Backend_Flush_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Backend_Flush_Call) RunAndReturn(run func() error) *Backend_Flush_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: key
func (_m *Backend) Get(key string) (httpcache.Entry, bool) {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 httpcache.Entry
	var r1 bool
	if rf, ok := ret.Get(0).(func(string) (httpcache.Entry, bool)); ok {
		return rf(key)
	}
	if rf, ok := ret.Get(0).(func(string) httpcache.Entry); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Get(0).(httpcache.Entry)
	}

	if rf, ok := ret.Get(1).(func(string) bool); ok {
		r1 = rf(key)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// Backend_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type Backend_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - key string
func (_e *Backend_Expecter) Get(key interface{}) *Backend_Get_Call {
	return &Backend_Get_Call{Call: _e.mock.On("Get", key)}
}

func (_c *Backend_Get_Call) Run(run func(key string)) *Backend_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Backend_Get_Call) Return(_a0 httpcache.Entry, _a1 bool) *Backend_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Backend_Get_Call) RunAndReturn(run func(string) (httpcache.Entry, bool)) *Backend_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Purge provides a mock function with given fields: key
func (_m *Backend) Purge(key string) error {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for Purge")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Backend_Purge_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Purge'
type Backend_Purge_Call struct {
	*mock.Call
}

// Purge is a helper method to define mock.On call
//   - key string
func (_e *Backend_Expecter) Purge(key interface{}) *Backend_Purge_Call {
	return &Backend_Purge_Call{Call: _e.mock.On("Purge", key)}
}

func (_c *Backend_Purge_Call) Run(run func(key string)) *Backend_Purge_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Backend_Purge_Call) Return(_a0 error) *Backend_Purge_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Backend_Purge_Call) RunAndReturn(run func(string) error) *Backend_Purge_Call {
	_c.Call.Return(run)
	return _c
}

// Set provides a mock function with given fields: key, entry
func (_m *Backend) Set(key string, entry httpcache.Entry) error {
	ret := _m.Called(key, entry)

	if len(ret) == 0 {
		panic("no return value specified for Set")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, httpcache.Entry) error); ok {
		r0 = rf(key, entry)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Backend_Set_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Set'
type Backend_Set_Call struct {
	*mock.Call
}

// Set is a helper method to define mock.On call
//   - key string
//   - entry httpcache.Entry
func (_e *Backend_Expecter) Set(key interface{}, entry interface{}) *Backend_Set_Call {
	return &Backend_Set_Call{Call: _e.mock.On("Set", key, entry)}
}

func (_c *Backend_Set_Call) Run(run func(key string, entry httpcache.Entry)) *Backend_Set_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(httpcache.Entry))
	})
	return _c
}

func (_c *Backend_Set_Call) Return(_a0 error) *Backend_Set_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Backend_Set_Call) RunAndReturn(run func(string, httpcache.Entry) error) *Backend_Set_Call {
	_c.Call.Return(run)
	return _c
}

// NewBackend creates a new instance of Backend. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBackend(t interface {
	mock.TestingT
	Cleanup(func())
}) *Backend {
	mock := &Backend{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
