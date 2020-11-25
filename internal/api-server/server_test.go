package apiserver

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunServer(t *testing.T) {
	errCh := make(chan error)
	stopCh := make(chan struct{})
	diagnosticsPort := 9999

	srv := newApiServer(diagnosticsPort, []string{}, nil)

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
}

func TestRunServerWithInvalidPort(t *testing.T) {
	errCh := make(chan error)
	stopCh := make(chan struct{})

	srv := newApiServer(-9999, []string{}, nil)

	go func() {
		defer close(errCh)
		errCh <- srv.Start(stopCh)
	}()

	assert.Error(t, <-errCh)
}
