// Code generated by MockGen. DO NOT EDIT.
// Source: service.go
//
// Generated by this command:
//
//	mockgen -source=service.go -destination=mock/service.go
//

// Package mock_sentence is a generated GoMock package.
package mock_sentence

import (
	reflect "reflect"

	openai "github.com/Lionel-Wilson/My-Language-Aibou-API/internal/clients/open-ai"
	gomock "go.uber.org/mock/gomock"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
	isgomock struct{}
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

// GetSentenceCorrection mocks base method.
func (m *MockService) GetSentenceCorrection(sentence, nativeLanguage string) (*openai.ChatCompletion, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSentenceCorrection", sentence, nativeLanguage)
	ret0, _ := ret[0].(*openai.ChatCompletion)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSentenceCorrection indicates an expected call of GetSentenceCorrection.
func (mr *MockServiceMockRecorder) GetSentenceCorrection(sentence, nativeLanguage any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSentenceCorrection", reflect.TypeOf((*MockService)(nil).GetSentenceCorrection), sentence, nativeLanguage)
}

// GetSentenceExplanation mocks base method.
func (m *MockService) GetSentenceExplanation(sentence, nativeLanguage string) (*openai.ChatCompletion, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSentenceExplanation", sentence, nativeLanguage)
	ret0, _ := ret[0].(*openai.ChatCompletion)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSentenceExplanation indicates an expected call of GetSentenceExplanation.
func (mr *MockServiceMockRecorder) GetSentenceExplanation(sentence, nativeLanguage any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSentenceExplanation", reflect.TypeOf((*MockService)(nil).GetSentenceExplanation), sentence, nativeLanguage)
}

// ValidateSentence mocks base method.
func (m *MockService) ValidateSentence(sentence string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateSentence", sentence)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateSentence indicates an expected call of ValidateSentence.
func (mr *MockServiceMockRecorder) ValidateSentence(sentence any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateSentence", reflect.TypeOf((*MockService)(nil).ValidateSentence), sentence)
}
