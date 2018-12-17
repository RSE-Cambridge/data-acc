// Code generated by MockGen. DO NOT EDIT.
// Source: internal/pkg/keystoreregistry/keystore.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	keystoreregistry "github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockKeystore is a mock of Keystore interface
type MockKeystore struct {
	ctrl     *gomock.Controller
	recorder *MockKeystoreMockRecorder
}

// MockKeystoreMockRecorder is the mock recorder for MockKeystore
type MockKeystoreMockRecorder struct {
	mock *MockKeystore
}

// NewMockKeystore creates a new mock instance
func NewMockKeystore(ctrl *gomock.Controller) *MockKeystore {
	mock := &MockKeystore{ctrl: ctrl}
	mock.recorder = &MockKeystoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockKeystore) EXPECT() *MockKeystoreMockRecorder {
	return m.recorder
}

// Close mocks base method
func (m *MockKeystore) Close() error {
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (mr *MockKeystoreMockRecorder) Close() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockKeystore)(nil).Close))
}

// CleanPrefix mocks base method
func (m *MockKeystore) CleanPrefix(prefix string) error {
	ret := m.ctrl.Call(m, "CleanPrefix", prefix)
	ret0, _ := ret[0].(error)
	return ret0
}

// CleanPrefix indicates an expected call of CleanPrefix
func (mr *MockKeystoreMockRecorder) CleanPrefix(prefix interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CleanPrefix", reflect.TypeOf((*MockKeystore)(nil).CleanPrefix), prefix)
}

// Add mocks base method
func (m *MockKeystore) Add(keyValues []keystoreregistry.KeyValue) error {
	ret := m.ctrl.Call(m, "Add", keyValues)
	ret0, _ := ret[0].(error)
	return ret0
}

// Add indicates an expected call of Add
func (mr *MockKeystoreMockRecorder) Add(keyValues interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockKeystore)(nil).Add), keyValues)
}

// Update mocks base method
func (m *MockKeystore) Update(keyValues []keystoreregistry.KeyValueVersion) error {
	ret := m.ctrl.Call(m, "Update", keyValues)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockKeystoreMockRecorder) Update(keyValues interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockKeystore)(nil).Update), keyValues)
}

// DeleteAll mocks base method
func (m *MockKeystore) DeleteAll(keyValues []keystoreregistry.KeyValueVersion) error {
	ret := m.ctrl.Call(m, "DeleteAll", keyValues)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAll indicates an expected call of DeleteAll
func (mr *MockKeystoreMockRecorder) DeleteAll(keyValues interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAll", reflect.TypeOf((*MockKeystore)(nil).DeleteAll), keyValues)
}

// GetAll mocks base method
func (m *MockKeystore) GetAll(prefix string) ([]keystoreregistry.KeyValueVersion, error) {
	ret := m.ctrl.Call(m, "GetAll", prefix)
	ret0, _ := ret[0].([]keystoreregistry.KeyValueVersion)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll
func (mr *MockKeystoreMockRecorder) GetAll(prefix interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockKeystore)(nil).GetAll), prefix)
}

// Get mocks base method
func (m *MockKeystore) Get(key string) (keystoreregistry.KeyValueVersion, error) {
	ret := m.ctrl.Call(m, "Get", key)
	ret0, _ := ret[0].(keystoreregistry.KeyValueVersion)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockKeystoreMockRecorder) Get(key interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockKeystore)(nil).Get), key)
}

// WatchPrefix mocks base method
func (m *MockKeystore) WatchPrefix(prefix string, onUpdate func(*keystoreregistry.KeyValueVersion, *keystoreregistry.KeyValueVersion)) {
	m.ctrl.Call(m, "WatchPrefix", prefix, onUpdate)
}

// WatchPrefix indicates an expected call of WatchPrefix
func (mr *MockKeystoreMockRecorder) WatchPrefix(prefix, onUpdate interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WatchPrefix", reflect.TypeOf((*MockKeystore)(nil).WatchPrefix), prefix, onUpdate)
}

// WatchKey mocks base method
func (m *MockKeystore) WatchKey(ctxt context.Context, key string, onUpdate func(*keystoreregistry.KeyValueVersion, *keystoreregistry.KeyValueVersion)) {
	m.ctrl.Call(m, "WatchKey", ctxt, key, onUpdate)
}

// WatchKey indicates an expected call of WatchKey
func (mr *MockKeystoreMockRecorder) WatchKey(ctxt, key, onUpdate interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WatchKey", reflect.TypeOf((*MockKeystore)(nil).WatchKey), ctxt, key, onUpdate)
}

// KeepAliveKey mocks base method
func (m *MockKeystore) KeepAliveKey(key string) error {
	ret := m.ctrl.Call(m, "KeepAliveKey", key)
	ret0, _ := ret[0].(error)
	return ret0
}

// KeepAliveKey indicates an expected call of KeepAliveKey
func (mr *MockKeystoreMockRecorder) KeepAliveKey(key interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "KeepAliveKey", reflect.TypeOf((*MockKeystore)(nil).KeepAliveKey), key)
}

// NewMutex mocks base method
func (m *MockKeystore) NewMutex(lockKey string) (keystoreregistry.Mutex, error) {
	ret := m.ctrl.Call(m, "NewMutex", lockKey)
	ret0, _ := ret[0].(keystoreregistry.Mutex)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewMutex indicates an expected call of NewMutex
func (mr *MockKeystoreMockRecorder) NewMutex(lockKey interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewMutex", reflect.TypeOf((*MockKeystore)(nil).NewMutex), lockKey)
}

// MockMutex is a mock of Mutex interface
type MockMutex struct {
	ctrl     *gomock.Controller
	recorder *MockMutexMockRecorder
}

// MockMutexMockRecorder is the mock recorder for MockMutex
type MockMutexMockRecorder struct {
	mock *MockMutex
}

// NewMockMutex creates a new mock instance
func NewMockMutex(ctrl *gomock.Controller) *MockMutex {
	mock := &MockMutex{ctrl: ctrl}
	mock.recorder = &MockMutexMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMutex) EXPECT() *MockMutexMockRecorder {
	return m.recorder
}

// Lock mocks base method
func (m *MockMutex) Lock(ctx context.Context) error {
	ret := m.ctrl.Call(m, "Lock", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Lock indicates an expected call of Lock
func (mr *MockMutexMockRecorder) Lock(ctx interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Lock", reflect.TypeOf((*MockMutex)(nil).Lock), ctx)
}

// Unlock mocks base method
func (m *MockMutex) Unlock(ctx context.Context) error {
	ret := m.ctrl.Call(m, "Unlock", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Unlock indicates an expected call of Unlock
func (mr *MockMutexMockRecorder) Unlock(ctx interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unlock", reflect.TypeOf((*MockMutex)(nil).Unlock), ctx)
}
