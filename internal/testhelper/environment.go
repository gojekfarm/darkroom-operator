package testhelper

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-logr/logr"
	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/gojekfarm/darkroom-operator/internal/controllers/setup"
)

type Environment interface {
	Start() error
	Stop() error
	Add(reconciler reconcile.Reconciler)
	GetConfig() *rest.Config
	GetLogs() string
	GetLogger() logr.Logger
	ResetLogs()
}

type env struct {
	k8sEnv      *envtest.Environment
	cfg         *rest.Config
	logger      logr.Logger
	buf         *bytes.Buffer
	stopCh      chan struct{}
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

	if len(e.reconcilers) > 0 {
		e.mgr, err = ctrl.NewManager(c, ctrl.Options{
			Scheme:             Scheme,
			CertDir:            e.k8sEnv.WebhookInstallOptions.LocalServingCertDir,
			Host:               e.k8sEnv.WebhookInstallOptions.LocalServingHost,
			Port:               e.k8sEnv.WebhookInstallOptions.LocalServingPort,
			MetricsBindAddress: fmt.Sprintf(":%d", freePort()),
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
			err := e.mgr.Start(e.stopCh)
			if err != nil {
				errCh <- err
			}
		}()
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			if strings.Contains(e.buf.String(), "Starting workers") {
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
	close(e.stopCh)
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

func NewTestEnvironment(CRDDirectoryPaths []string) Environment {
	b := &bytes.Buffer{}
	l := zap.New(zap.UseDevMode(true), zap.WriteTo(b))
	logf.SetLogger(l)

	return &env{
		k8sEnv: &envtest.Environment{
			CRDDirectoryPaths:     CRDDirectoryPaths,
			ErrorIfCRDPathMissing: true,
			WebhookInstallOptions: envtest.WebhookInstallOptions{
				MutatingWebhooks: getMutationWebhooks(),
			},
		},
		reconcilers: []reconcile.Reconciler{},
		buf:         b,
		logger:      l,
		stopCh:      make(chan struct{}),
	}
}

func getMutationWebhooks() []runtime.Object {
	failedTypeV1Beta1 := admissionv1beta1.Fail
	webhookPathV1 := "/mutate-deployments-gojek-io-v1alpha1-darkroom"

	return []runtime.Object{
		&admissionv1beta1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "mutating-webhook-configuration",
			},
			TypeMeta: metav1.TypeMeta{
				Kind:       "MutatingWebhookConfiguration",
				APIVersion: "admissionregistration.k8s.io/v1beta1",
			},
			Webhooks: []admissionv1beta1.MutatingWebhook{
				{
					Name: "mdarkroom.gojek.io",
					ClientConfig: admissionv1beta1.WebhookClientConfig{
						Service: &admissionv1beta1.ServiceReference{
							Name:      "webhook-service",
							Namespace: "system",
							Path:      &webhookPathV1,
						},
					},
					Rules: []admissionv1beta1.RuleWithOperations{
						{
							Operations: []admissionv1beta1.OperationType{"CREATE", "UPDATE"},
							Rule: admissionv1beta1.Rule{
								APIGroups:   []string{"deployments.gojek.io"},
								APIVersions: []string{"v1alpha1"},
								Resources:   []string{"darkrooms"},
							},
						},
					},
					FailurePolicy: &failedTypeV1Beta1,
				},
			},
		},
	}
}

func freePort() int {
	addr, _ := net.ResolveTCPAddr("tcp", "localhost:0")
	l, _ := net.ListenTCP("tcp", addr)
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
