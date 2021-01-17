// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/michael-diggin/yass/proto (interfaces: YassServiceClient)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	proto "github.com/michael-diggin/yass/proto"
	grpc "google.golang.org/grpc"
	reflect "reflect"
)

// MockYassServiceClient is a mock of YassServiceClient interface
type MockYassServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockYassServiceClientMockRecorder
}

// MockYassServiceClientMockRecorder is the mock recorder for MockYassServiceClient
type MockYassServiceClientMockRecorder struct {
	mock *MockYassServiceClient
}

// NewMockYassServiceClient creates a new mock instance
func NewMockYassServiceClient(ctrl *gomock.Controller) *MockYassServiceClient {
	mock := &MockYassServiceClient{ctrl: ctrl}
	mock.recorder = &MockYassServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockYassServiceClient) EXPECT() *MockYassServiceClientMockRecorder {
	return m.recorder
}

// Put mocks base method
func (m *MockYassServiceClient) Put(arg0 context.Context, arg1 *proto.Pair, arg2 ...grpc.CallOption) (*proto.Null, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Put", varargs...)
	ret0, _ := ret[0].(*proto.Null)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Put indicates an expected call of Put
func (mr *MockYassServiceClientMockRecorder) Put(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Put", reflect.TypeOf((*MockYassServiceClient)(nil).Put), varargs...)
}

// Retrieve mocks base method
func (m *MockYassServiceClient) Retrieve(arg0 context.Context, arg1 *proto.Key, arg2 ...grpc.CallOption) (*proto.Pair, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Retrieve", varargs...)
	ret0, _ := ret[0].(*proto.Pair)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Retrieve indicates an expected call of Retrieve
func (mr *MockYassServiceClientMockRecorder) Retrieve(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Retrieve", reflect.TypeOf((*MockYassServiceClient)(nil).Retrieve), varargs...)
}