package cmd

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
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

	mm := &mockManager{}
	mm.On("Start", mock.AnythingOfType("<-chan struct {}")).Return(nil)

	s.rootCmd = newRootCmd(rootCmdOpts{
		SetupSignalHandler: func() <-chan struct{} {
			return s.stopCh
		},
		NewManager: func(config *rest.Config, options ctrl.Options) (ctrl.Manager, error) {
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

	s.True(s.Eventually(func() bool {
		return strings.Contains(s.buf.String(), "starting manager")
	}, 5*time.Second, 100*time.Millisecond), "failed to start controller")

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
		NewManager: func(config *rest.Config, options ctrl.Options) (ctrl.Manager, error) {
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
		NewManager: func(config *rest.Config, options ctrl.Options) (ctrl.Manager, error) {
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

func (m *mockManager) Add(runnable manager.Runnable) error {
	return m.Called(runnable).Error(0)
}

func (m *mockManager) Elected() <-chan struct{} {
	return m.Called().Get(0).(<-chan struct{})
}

func (m *mockManager) SetFields(i interface{}) error {
	return m.Called(i).Error(0)
}

func (m *mockManager) AddMetricsExtraHandler(path string, handler http.Handler) error {
	return m.Called(path, handler).Error(0)
}

func (m *mockManager) AddHealthzCheck(name string, check healthz.Checker) error {
	return m.Called(name, check).Error(0)
}

func (m *mockManager) AddReadyzCheck(name string, check healthz.Checker) error {
	return m.Called(name, check).Error(0)
}

func (m *mockManager) Start(i <-chan struct{}) error {
	return m.Called(i).Error(0)
}

func (m *mockManager) GetConfig() *rest.Config {
	return m.Called().Get(0).(*rest.Config)
}

func (m *mockManager) GetScheme() *runtime.Scheme {
	return m.Called().Get(0).(*runtime.Scheme)
}

func (m *mockManager) GetClient() client.Client {
	return m.Called().Get(0).(client.Client)
}

func (m *mockManager) GetFieldIndexer() client.FieldIndexer {
	return m.Called().Get(0).(client.FieldIndexer)
}

func (m *mockManager) GetCache() cache.Cache {
	return m.Called().Get(0).(cache.Cache)
}

func (m *mockManager) GetEventRecorderFor(name string) record.EventRecorder {
	return m.Called(name).Get(0).(record.EventRecorder)
}

func (m *mockManager) GetRESTMapper() meta.RESTMapper {
	return m.Called().Get(0).(meta.RESTMapper)
}

func (m *mockManager) GetAPIReader() client.Reader {
	return m.Called().Get(0).(client.Reader)
}

func (m *mockManager) GetWebhookServer() *webhook.Server {
	return m.Called().Get(0).(*webhook.Server)
}

func (m *mockManager) GetLogger() logr.Logger {
	return m.Called().Get(0).(logr.Logger)
}
