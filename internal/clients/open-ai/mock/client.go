// Code generated by MockGen. DO NOT EDIT.
// Source: client.go
//
// Generated by this command:
//
//	mockgen -source=client.go -destination=mock/client.go
//

// Package mock_openai is a generated GoMock package.
package mock_openai

import (
	http "net/http"
	reflect "reflect"
	strings "strings"

	gomock "go.uber.org/mock/gomock"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// MakeRequest mocks base method.
func (m *MockClient) MakeRequest(body *strings.Reader) (*http.Response, []byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MakeRequest", body)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].([]byte)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// MakeRequest indicates an expected call of MakeRequest.
func (mr *MockClientMockRecorder) MakeRequest(body any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MakeRequest", reflect.TypeOf((*MockClient)(nil).MakeRequest), body)
}
