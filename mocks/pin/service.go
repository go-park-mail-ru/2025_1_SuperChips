// Code generated by MockGen. DO NOT EDIT.
// Source: ./pin/service.go
//
// Generated by this command:
//
//	mockgen -source=./pin/service.go -destination=./mocks/pin/service.go
//

// Package mock_pin is a generated GoMock package.
package mock_pin

import (
	reflect "reflect"

	domain "github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	gomock "go.uber.org/mock/gomock"
)

// MockPinRepository is a mock of PinRepository interface.
type MockPinRepository struct {
	ctrl     *gomock.Controller
	recorder *MockPinRepositoryMockRecorder
	isgomock struct{}
}

// MockPinRepositoryMockRecorder is the mock recorder for MockPinRepository.
type MockPinRepositoryMockRecorder struct {
	mock *MockPinRepository
}

// NewMockPinRepository creates a new mock instance.
func NewMockPinRepository(ctrl *gomock.Controller) *MockPinRepository {
	mock := &MockPinRepository{ctrl: ctrl}
	mock.recorder = &MockPinRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPinRepository) EXPECT() *MockPinRepositoryMockRecorder {
	return m.recorder
}

// GetPins mocks base method.
func (m *MockPinRepository) GetPins(page, pageSize int) ([]domain.PinData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPins", page, pageSize)
	ret0, _ := ret[0].([]domain.PinData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPins indicates an expected call of GetPins.
func (mr *MockPinRepositoryMockRecorder) GetPins(page, pageSize any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPins", reflect.TypeOf((*MockPinRepository)(nil).GetPins), page, pageSize)
}
