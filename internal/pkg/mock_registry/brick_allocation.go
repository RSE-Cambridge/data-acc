// Code generated by MockGen. DO NOT EDIT.
// Source: internal/pkg/registry/brick_allocation.go

// Package mock_registry is a generated GoMock package.
package mock_registry

import (
	datamodel "github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	store "github.com/RSE-Cambridge/data-acc/internal/pkg/store"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockAllocationRegistry is a mock of AllocationRegistry interface
type MockAllocationRegistry struct {
	ctrl     *gomock.Controller
	recorder *MockAllocationRegistryMockRecorder
}

// MockAllocationRegistryMockRecorder is the mock recorder for MockAllocationRegistry
type MockAllocationRegistryMockRecorder struct {
	mock *MockAllocationRegistry
}

// NewMockAllocationRegistry creates a new mock instance
func NewMockAllocationRegistry(ctrl *gomock.Controller) *MockAllocationRegistry {
	mock := &MockAllocationRegistry{ctrl: ctrl}
	mock.recorder = &MockAllocationRegistryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAllocationRegistry) EXPECT() *MockAllocationRegistryMockRecorder {
	return m.recorder
}

// GetAllocationMutex mocks base method
func (m *MockAllocationRegistry) GetAllocationMutex() (store.Mutex, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllocationMutex")
	ret0, _ := ret[0].(store.Mutex)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllocationMutex indicates an expected call of GetAllocationMutex
func (mr *MockAllocationRegistryMockRecorder) GetAllocationMutex() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllocationMutex", reflect.TypeOf((*MockAllocationRegistry)(nil).GetAllocationMutex))
}

// GetPool mocks base method
func (m *MockAllocationRegistry) GetPool(name datamodel.PoolName) (datamodel.Pool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPool", name)
	ret0, _ := ret[0].(datamodel.Pool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPool indicates an expected call of GetPool
func (mr *MockAllocationRegistryMockRecorder) GetPool(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPool", reflect.TypeOf((*MockAllocationRegistry)(nil).GetPool), name)
}

// EnsurePoolCreated mocks base method
func (m *MockAllocationRegistry) EnsurePoolCreated(poolName datamodel.PoolName, granularityBytes uint) (datamodel.Pool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsurePoolCreated", poolName, granularityBytes)
	ret0, _ := ret[0].(datamodel.Pool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsurePoolCreated indicates an expected call of EnsurePoolCreated
func (mr *MockAllocationRegistryMockRecorder) EnsurePoolCreated(poolName, granularityBytes interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsurePoolCreated", reflect.TypeOf((*MockAllocationRegistry)(nil).EnsurePoolCreated), poolName, granularityBytes)
}

// GetAllPoolInfos mocks base method
func (m *MockAllocationRegistry) GetAllPoolInfos() ([]datamodel.PoolInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllPoolInfos")
	ret0, _ := ret[0].([]datamodel.PoolInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllPoolInfos indicates an expected call of GetAllPoolInfos
func (mr *MockAllocationRegistryMockRecorder) GetAllPoolInfos() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllPoolInfos", reflect.TypeOf((*MockAllocationRegistry)(nil).GetAllPoolInfos))
}

// GetPoolInfo mocks base method
func (m *MockAllocationRegistry) GetPoolInfo(poolName datamodel.PoolName) (datamodel.PoolInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPoolInfo", poolName)
	ret0, _ := ret[0].(datamodel.PoolInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPoolInfo indicates an expected call of GetPoolInfo
func (mr *MockAllocationRegistryMockRecorder) GetPoolInfo(poolName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPoolInfo", reflect.TypeOf((*MockAllocationRegistry)(nil).GetPoolInfo), poolName)
}
