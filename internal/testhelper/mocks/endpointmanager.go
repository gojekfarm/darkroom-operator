package mocks

import (
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/mock"
)

type MockEndpointManager struct {
	mock.Mock
}

func (m *MockEndpointManager) Setup(c *restful.Container) {
	m.Called(c)
}
