// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/michael-diggin/yass/server/model (interfaces: Service)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	model "github.com/michael-diggin/yass/server/model"
	reflect "reflect"
)

// MockService is a mock of Service interface
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// BatchGet mocks base method
func (m *MockService) BatchGet() <-chan map[string]model.Data {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchGet")
	ret0, _ := ret[0].(<-chan map[string]model.Data)
	return ret0
}

// BatchGet indicates an expected call of BatchGet
func (mr *MockServiceMockRecorder) BatchGet() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchGet", reflect.TypeOf((*MockService)(nil).BatchGet))
}

// BatchSet mocks base method
func (m *MockService) BatchSet(arg0 map[string]model.Data) <-chan error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BatchSet", arg0)
	ret0, _ := ret[0].(<-chan error)
	return ret0
}

// BatchSet indicates an expected call of BatchSet
func (mr *MockServiceMockRecorder) BatchSet(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BatchSet", reflect.TypeOf((*MockService)(nil).BatchSet), arg0)
}

// Close mocks base method
func (m *MockService) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close
func (mr *MockServiceMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockService)(nil).Close))
}

// Delete mocks base method
func (m *MockService) Delete(arg0 string) <-chan *model.StorageResponse {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0)
	ret0, _ := ret[0].(<-chan *model.StorageResponse)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockServiceMockRecorder) Delete(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockService)(nil).Delete), arg0)
}

// Get mocks base method
func (m *MockService) Get(arg0 string) <-chan *model.StorageResponse {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(<-chan *model.StorageResponse)
	return ret0
}

// Get indicates an expected call of Get
func (mr *MockServiceMockRecorder) Get(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockService)(nil).Get), arg0)
}

// Ping mocks base method
func (m *MockService) Ping() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping
func (mr *MockServiceMockRecorder) Ping() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockService)(nil).Ping))
}

// Set mocks base method
func (m *MockService) Set(arg0 string, arg1 uint32, arg2 interface{}) <-chan *model.StorageResponse {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", arg0, arg1, arg2)
	ret0, _ := ret[0].(<-chan *model.StorageResponse)
	return ret0
}

// Set indicates an expected call of Set
func (mr *MockServiceMockRecorder) Set(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockService)(nil).Set), arg0, arg1, arg2)
}
