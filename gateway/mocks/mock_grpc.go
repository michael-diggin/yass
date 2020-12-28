// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/michael-diggin/yass/gateway/api (interfaces: GrpcClient)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	models "github.com/michael-diggin/yass/models"
	reflect "reflect"
)

// MockGrpcClient is a mock of GrpcClient interface
type MockGrpcClient struct {
	ctrl     *gomock.Controller
	recorder *MockGrpcClientMockRecorder
}

// MockGrpcClientMockRecorder is the mock recorder for MockGrpcClient
type MockGrpcClientMockRecorder struct {
	mock *MockGrpcClient
}

// NewMockGrpcClient creates a new mock instance
func NewMockGrpcClient(ctrl *gomock.Controller) *MockGrpcClient {
	mock := &MockGrpcClient{ctrl: ctrl}
	mock.recorder = &MockGrpcClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockGrpcClient) EXPECT() *MockGrpcClientMockRecorder {
	return m.recorder
}

// Close mocks base method
func (m *MockGrpcClient) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (mr *MockGrpcClientMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockGrpcClient)(nil).Close))
}

// DelFollowerValue mocks base method
func (m *MockGrpcClient) DelFollowerValue(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DelFollowerValue", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DelFollowerValue indicates an expected call of DelFollowerValue
func (mr *MockGrpcClientMockRecorder) DelFollowerValue(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DelFollowerValue", reflect.TypeOf((*MockGrpcClient)(nil).DelFollowerValue), arg0, arg1)
}

// DelValue mocks base method
func (m *MockGrpcClient) DelValue(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DelValue", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DelValue indicates an expected call of DelValue
func (mr *MockGrpcClientMockRecorder) DelValue(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DelValue", reflect.TypeOf((*MockGrpcClient)(nil).DelValue), arg0, arg1)
}

// GetFollowerValue mocks base method
func (m *MockGrpcClient) GetFollowerValue(arg0 context.Context, arg1 string) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFollowerValue", arg0, arg1)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFollowerValue indicates an expected call of GetFollowerValue
func (mr *MockGrpcClientMockRecorder) GetFollowerValue(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFollowerValue", reflect.TypeOf((*MockGrpcClient)(nil).GetFollowerValue), arg0, arg1)
}

// GetValue mocks base method
func (m *MockGrpcClient) GetValue(arg0 context.Context, arg1 string) (*models.Pair, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValue", arg0, arg1)
	ret0, _ := ret[0].(*models.Pair)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetValue indicates an expected call of GetValue
func (mr *MockGrpcClientMockRecorder) GetValue(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValue", reflect.TypeOf((*MockGrpcClient)(nil).GetValue), arg0, arg1)
}

// SetFollowerValue mocks base method
func (m *MockGrpcClient) SetFollowerValue(arg0 context.Context, arg1 string, arg2 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetFollowerValue", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetFollowerValue indicates an expected call of SetFollowerValue
func (mr *MockGrpcClientMockRecorder) SetFollowerValue(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetFollowerValue", reflect.TypeOf((*MockGrpcClient)(nil).SetFollowerValue), arg0, arg1, arg2)
}

// SetValue mocks base method
func (m *MockGrpcClient) SetValue(arg0 context.Context, arg1 *models.Pair) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetValue", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetValue indicates an expected call of SetValue
func (mr *MockGrpcClientMockRecorder) SetValue(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetValue", reflect.TypeOf((*MockGrpcClient)(nil).SetValue), arg0, arg1)
}