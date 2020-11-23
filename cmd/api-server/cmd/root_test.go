package cmd

import (
	"bytes"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/rest"

	apiserver "github.com/gojekfarm/darkroom-operator/internal/api-server"
)

type RootCmdSuite struct {
	suite.Suite
	rootCmd *cobra.Command
	buf     *bytes.Buffer
	stopCh  chan struct{}
}

func TestRootCmd(t *testing.T) {
	suite.Run(t, new(RootCmdSuite))
}

func (s *RootCmdSuite) SetupTest() {
	s.stopCh = make(chan struct{})
	s.buf = &bytes.Buffer{}
}

func (s *RootCmdSuite) TestNewRootCmd() {
	cmd := NewRootCmd()
	s.NotNil(cmd)
}

func (s *RootCmdSuite) TestControllerStartup() {
	errCh := make(chan error)
	defer close(errCh)
	var wg sync.WaitGroup
	wg.Add(1)

	mm := &mockManager{}

	mm.On("Start", mock.AnythingOfType("<-chan struct {}")).Return(nil)

	s.rootCmd = newRootCmd(rootCmdOpts{
		SetupSignalHandler: func() <-chan struct{} {
			time.AfterFunc(2*time.Second, func() {
				wg.Done()
			})
			return s.stopCh
		},
		NewManager: func(config *rest.Config, options apiserver.Options) (apiserver.Manager, error) {
			return mm, nil
		},
		GetConfigOrDie: func() *rest.Config {
			return nil
		},
	})
	s.rootCmd.SetArgs([]string{})
	s.rootCmd.SetOut(s.buf)

	go func() {
		errCh <- s.rootCmd.Execute()
	}()

	wg.Wait()
	close(s.stopCh)
	s.NoError(<-errCh)
}

func (s *RootCmdSuite) TestControllerStartupWithNewManagerError() {
	errCh := make(chan error)
	managerErr := errors.New("unable to create manager")

	s.rootCmd = newRootCmd(rootCmdOpts{
		SetupSignalHandler: func() <-chan struct{} {
			return s.stopCh
		},
		NewManager: func(config *rest.Config, options apiserver.Options) (apiserver.Manager, error) {
			return nil, managerErr
		},
		GetConfigOrDie: func() *rest.Config {
			return nil
		},
	})
	s.rootCmd.SetOut(s.buf)

	go func() {
		defer close(errCh)
		errCh <- s.rootCmd.Execute()
	}()

	close(s.stopCh)
	s.EqualError(<-errCh, managerErr.Error())
}

func (s *RootCmdSuite) TestControllerStartupWithManagerStartError() {
	errCh := make(chan error)
	startErr := errors.New("unable to start manager")

	mm := &mockManager{}

	mm.On("Start", mock.AnythingOfType("<-chan struct {}")).Return(startErr)

	s.rootCmd = newRootCmd(rootCmdOpts{
		SetupSignalHandler: func() <-chan struct{} {
			return s.stopCh
		},
		NewManager: func(config *rest.Config, options apiserver.Options) (apiserver.Manager, error) {
			return mm, nil
		},
		GetConfigOrDie: func() *rest.Config {
			return nil
		},
	})
	s.rootCmd.SetOut(s.buf)

	go func() {
		defer close(errCh)
		errCh <- s.rootCmd.Execute()
	}()

	close(s.stopCh)
	s.EqualError(<-errCh, startErr.Error())
}

type mockManager struct {
	mock.Mock
}

func (m *mockManager) Start(stopCh <-chan struct{}) error {
	return m.Called(stopCh).Error(0)
}
