package cmd

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"sync"
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
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	internalRuntime "github.com/gojekfarm/darkroom-operator/internal/runtime"
	"github.com/gojekfarm/darkroom-operator/internal/testhelper/mocks"
)

type RootCmdSuite struct {
	suite.Suite
	rootCmd   *cobra.Command
	buf       *bytes.Buffer
	ctx       context.Context
	cancelCtx func()
}

func TestRootCmd(t *testing.T) {
	suite.Run(t, new(RootCmdSuite))
}

func (s *RootCmdSuite) SetupTest() {
	s.ctx, s.cancelCtx = context.WithCancel(context.Background())
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
	mc := &mocks.MockClient{RuntimeScheme: internalRuntime.Scheme()}

	mm.On("Start", mock.AnythingOfType("*context.cancelCtx")).Return(nil)
	mm.On("GetClient").Return(mc)
	mm.On("GetScheme").Return(internalRuntime.Scheme())
	mm.On("GetConfig").Return(&rest.Config{})
	mm.On("GetLogger").Return(zap.New(zap.UseDevMode(true)))
	mm.On("SetFields", mock.Anything).Return(nil)
	mm.On("AddHealthzCheck", "healthz", mock.AnythingOfType("healthz.Checker")).Return(nil)
	mm.On("AddReadyzCheck", "readyz", mock.AnythingOfType("healthz.Checker")).Return(nil)
	mm.On("Add", mock.AnythingOfType("*controller.Controller")).Return(nil)

	s.rootCmd = newRootCmd(rootCmdOpts{
		SetupSignalHandler: func() context.Context {
			time.AfterFunc(2*time.Second, func() {
				wg.Done()
			})
			return s.ctx
		},
		NewManager: func(config *rest.Config, options ctrl.Options) (ctrl.Manager, error) {
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
	s.cancelCtx()
	s.NoError(<-errCh)
}

func (s *RootCmdSuite) TestControllerStartupWithNewManagerError() {
	errCh := make(chan error)
	managerErr := errors.New("unable to create manager")

	s.rootCmd = newRootCmd(rootCmdOpts{
		SetupSignalHandler: func() context.Context {
			return s.ctx
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

	s.cancelCtx()
	s.EqualError(<-errCh, managerErr.Error())
}

func (s *RootCmdSuite) TestControllerStartupWithManagerStartError() {
	errCh := make(chan error)
	startErr := errors.New("unable to start manager")

	mm := &mockManager{}
	mc := &mocks.MockClient{RuntimeScheme: internalRuntime.Scheme()}

	mm.On("Start", mock.AnythingOfType("*context.cancelCtx")).Return(startErr)
	mm.On("GetClient").Return(mc)
	mm.On("GetScheme").Return(internalRuntime.Scheme())
	mm.On("GetConfig").Return(&rest.Config{})
	mm.On("GetLogger").Return(zap.New(zap.UseDevMode(true)))
	mm.On("SetFields", mock.Anything).Return(nil)
	mm.On("AddHealthzCheck", "healthz", mock.AnythingOfType("healthz.Checker")).Return(nil)
	mm.On("AddReadyzCheck", "readyz", mock.AnythingOfType("healthz.Checker")).Return(nil)
	mm.On("Add", mock.AnythingOfType("*controller.Controller")).Return(nil)

	s.rootCmd = newRootCmd(rootCmdOpts{
		SetupSignalHandler: func() context.Context {
			return s.ctx
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

	s.cancelCtx()
	s.EqualError(<-errCh, startErr.Error())
}

func (s *RootCmdSuite) TestControllerSetupError() {
	errCh := make(chan error)
	controllerAddErr := errors.New("can't add controller")

	mm := &mockManager{}
	mc := &mocks.MockClient{RuntimeScheme: internalRuntime.Scheme()}

	mm.On("Start", mock.AnythingOfType("*context.cancelCtx")).Return(nil)
	mm.On("GetClient").Return(mc)
	mm.On("GetScheme").Return(internalRuntime.Scheme())
	mm.On("GetConfig").Return(&rest.Config{})
	mm.On("GetLogger").Return(zap.New(zap.UseDevMode(true)))
	mm.On("SetFields", mock.Anything).Return(nil)
	mm.On("AddHealthzCheck", "healthz", mock.AnythingOfType("healthz.Checker")).Return(nil)
	mm.On("AddReadyzCheck", "readyz", mock.AnythingOfType("healthz.Checker")).Return(nil)
	mm.On("Add", mock.AnythingOfType("*controller.Controller")).Return(controllerAddErr)

	s.rootCmd = newRootCmd(rootCmdOpts{
		SetupSignalHandler: func() context.Context {
			return s.ctx
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

	s.cancelCtx()
	s.EqualError(<-errCh, controllerAddErr.Error())
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

func (m *mockManager) Start(ctx context.Context) error {
	return m.Called(ctx).Error(0)
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
	return &webhook.Server{}
}

func (m *mockManager) GetLogger() logr.Logger {
	return m.Called().Get(0).(logr.Logger)
}
