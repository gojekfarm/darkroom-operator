package apiserver

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gojekfarm/darkroom-operator/internal/testhelper/mocks"
)

func TestRunServer(t *testing.T) {
	errCh := make(chan error)
	stopCh := make(chan struct{})
	diagnosticsPort := 9999
	em := &mocks.MockEndpointManager{}
	em.On("Setup", mock.AnythingOfType("*restful.Container"))

	srv := newApiServer(diagnosticsPort, []string{}, em)

	go func() {
		defer close(errCh)
		errCh <- srv.Start(stopCh)
	}()

	assert.True(t, assert.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/healthz", diagnosticsPort))
		if err != nil {
			return false
		}
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond), "failed to run server")

	close(stopCh)

	assert.NoError(t, <-errCh)
	em.AssertExpectations(t)
}

func TestRunServerWithInvalidPort(t *testing.T) {
	errCh := make(chan error)
	stopCh := make(chan struct{})
	em := &mocks.MockEndpointManager{}
	em.On("Setup", mock.AnythingOfType("*restful.Container"))

	srv := newApiServer(-9999, []string{}, em)

	go func() {
		defer close(errCh)
		errCh <- srv.Start(stopCh)
	}()

	assert.Error(t, <-errCh)
	em.AssertExpectations(t)
}
