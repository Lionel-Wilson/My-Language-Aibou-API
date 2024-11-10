// Code generated by MockGen. DO NOT EDIT.
// Source: service.go
//
// Generated by this command:
//
//	mockgen -source=service.go -destination=mock/service.go
//

// Package mock_word is a generated GoMock package.
package mock_word

import (
	reflect "reflect"

	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
	gin "github.com/gin-gonic/gin"
	gomock "go.uber.org/mock/gomock"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// GetWordDefinition mocks base method.
func (m *MockService) GetWordDefinition(c *gin.Context, word, nativeLanguage string) (*openai.ChatCompletion, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWordDefinition", c, word, nativeLanguage)
	ret0, _ := ret[0].(*openai.ChatCompletion)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWordDefinition indicates an expected call of GetWordDefinition.
func (mr *MockServiceMockRecorder) GetWordDefinition(c, word, nativeLanguage any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWordDefinition", reflect.TypeOf((*MockService)(nil).GetWordDefinition), c, word, nativeLanguage)
}

// GetWordSynonyms mocks base method.
func (m *MockService) GetWordSynonyms(c *gin.Context, word, nativeLanguage string) (*openai.ChatCompletion, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWordSynonyms", c, word, nativeLanguage)
	ret0, _ := ret[0].(*openai.ChatCompletion)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWordSynonyms indicates an expected call of GetWordSynonyms.
func (mr *MockServiceMockRecorder) GetWordSynonyms(c, word, nativeLanguage any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWordSynonyms", reflect.TypeOf((*MockService)(nil).GetWordSynonyms), c, word, nativeLanguage)
}

// ValidateWord mocks base method.
func (m *MockService) ValidateWord(word string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateWord", word)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateWord indicates an expected call of ValidateWord.
func (mr *MockServiceMockRecorder) ValidateWord(word any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateWord", reflect.TypeOf((*MockService)(nil).ValidateWord), word)
}
