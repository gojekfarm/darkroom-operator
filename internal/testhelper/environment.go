package testhelper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/gojekfarm/darkroom-operator/internal/controllers/setup"
	"github.com/gojekfarm/darkroom-operator/internal/runtime"
)

type waitOptions struct {
	webhook bool
	worker  bool
}

type WaitOption interface {
	Apply(*waitOptions)
}

type Environment interface {
	Start() error
	Stop() error
	Add(reconciler reconcile.Reconciler)
	GetConfig() *rest.Config
	GetLogs() string
	GetLogger() logr.Logger
	ResetLogs()
	WithWaitOptions(...WaitOption) Environment
}

type WebhookWaitOption bool

func (w WebhookWaitOption) Apply(options *waitOptions) {
	options.webhook = bool(w)
}

type WorkerWaitOption bool

func (w WorkerWaitOption) Apply(options *waitOptions) {
	options.worker = bool(w)
}

type env struct {
	ctx         context.Context
	cancelCtx   func()
	waitOptions *waitOptions
	k8sEnv      *envtest.Environment
	cfg         *rest.Config
	logger      logr.Logger
	buf         *bytes.Buffer
	mgr         ctrl.Manager
	reconcilers []reconcile.Reconciler
}

func (e *env) Start() error {
	errCh := make(chan error)
	c, err := e.k8sEnv.Start()
	if err != nil {
		panic(err)
	}
	e.cfg = c

	e.mgr, err = ctrl.NewManager(c, ctrl.Options{
		Scheme:             runtime.Scheme(),
		CertDir:            e.k8sEnv.WebhookInstallOptions.LocalServingCertDir,
		Host:               e.k8sEnv.WebhookInstallOptions.LocalServingHost,
		Port:               e.k8sEnv.WebhookInstallOptions.LocalServingPort,
		MetricsBindAddress: fmt.Sprintf(":%d", FreePort()),
	})
	if err != nil {
		panic(err)
	}

	for _, r := range e.reconcilers {
		if s, ok := r.(setup.Controller); ok {
			if err := s.SetupControllerWithManager(e.mgr); err != nil {
				panic(err)
			}
		}

		if s, ok := r.(setup.Webhook); ok {
			if err := s.SetupWebhookWithManager(e.mgr); err != nil {
				panic(err)
			}
		}
	}
	go func() {
		err := e.mgr.Start(e.ctx)
		if err != nil {
			errCh <- err
		}
	}()

	if !e.waitOptions.webhook && !e.waitOptions.worker {
		return nil
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	webhookStarted := false
	workerStarted := false
	for {
		select {
		case <-ticker.C:
			if e.waitOptions.webhook && strings.Contains(e.buf.String(), "serving webhook server") {
				webhookStarted = true
			}
			if e.waitOptions.worker && strings.Contains(e.buf.String(), "Starting workers") {
				workerStarted = true
			}
			if e.waitOptions.webhook == webhookStarted && e.waitOptions.worker == workerStarted {
				ticker.Stop()
				return nil
			}
		case <-time.After(10 * time.Second):
			ticker.Stop()
			return errors.New("unable to start workers via controller manager")
		case err := <-errCh:
			return err
		}
	}
}

func (e *env) Add(reconciler reconcile.Reconciler) {
	e.reconcilers = append(e.reconcilers, reconciler)
}

func (e *env) Stop() error {
	e.cancelCtx()
	return e.k8sEnv.Stop()
}

func (e *env) GetConfig() *rest.Config {
	return e.cfg
}

func (e *env) GetLogger() logr.Logger {
	return e.logger
}

func (e *env) GetLogs() string {
	return e.buf.String()
}

func (e *env) ResetLogs() {
	e.buf.Reset()
}

func (e *env) WithWaitOptions(options ...WaitOption) Environment {
	for _, option := range options {
		option.Apply(e.waitOptions)
	}
	return e
}

func NewTestEnvironment(dirElems ...string) Environment {
	b := &bytes.Buffer{}
	l := zap.New(zap.UseDevMode(true), zap.WriteTo(b), zap.Level(zapcore.DebugLevel))
	logf.SetLogger(l)

	crdPaths := append(dirElems, "config", "crd", "bases")
	webhookPaths := append(dirElems, "config", "webhook")

	ctx, cancel := context.WithCancel(context.Background())

	return &env{
		ctx:       ctx,
		cancelCtx: cancel,
		waitOptions: &waitOptions{
			webhook: true,
			worker:  true,
		},
		k8sEnv: &envtest.Environment{
			CRDInstallOptions: envtest.CRDInstallOptions{
				Paths:              []string{filepath.Join(crdPaths...)},
				ErrorIfPathMissing: true,
			},
			WebhookInstallOptions: envtest.WebhookInstallOptions{
				Paths: []string{filepath.Join(webhookPaths...)},
			},
		},
		reconcilers: []reconcile.Reconciler{},
		buf:         b,
		logger:      l,
	}
}

func FreePort() int {
	addr, _ := net.ResolveTCPAddr("tcp", "localhost:0")
	l, _ := net.ListenTCP("tcp", addr)
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
