/*
MIT License

Copyright (c) 2020 GO-JEK Tech

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package controllers

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/gojekfarm/darkroom-operator/internal/testhelper"
	deploymentsv1alpha1 "github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

type DarkroomControllerSuite struct {
	suite.Suite
	cfg        *rest.Config
	k8sClient  client.Client
	testEnv    *envtest.Environment
	buf        *bytes.Buffer
	stopCh     chan struct{}
	mgr        ctrl.Manager
	reconciler *DarkroomReconciler
}

func TestDarkroomControllerSuite(t *testing.T) {
	suite.Run(t, new(DarkroomControllerSuite))
}

func (s *DarkroomControllerSuite) SetupSuite() {
	s.buf = &bytes.Buffer{}
	l := zap.New(zap.UseDevMode(true), zap.WriteTo(s.buf))
	logf.SetLogger(l)
	scheme := runtime.NewScheme()

	s.testEnv = testhelper.NewTestEnvironment()
	s.stopCh = make(chan struct{})

	var err error
	s.cfg, err = s.testEnv.Start()
	s.NoError(err)
	s.NotNil(s.cfg)

	err = clientgoscheme.AddToScheme(scheme)
	s.NoError(err)
	err = deploymentsv1alpha1.AddToScheme(scheme)
	s.NoError(err)

	// +kubebuilder:scaffold:scheme

	s.k8sClient, err = client.New(s.cfg, client.Options{Scheme: scheme})
	s.NoError(err)
	s.NotNil(s.k8sClient)

	s.mgr, err = ctrl.NewManager(s.cfg, ctrl.Options{
		Scheme:  scheme,
		CertDir: s.testEnv.WebhookInstallOptions.LocalServingCertDir,
		Host:    s.testEnv.WebhookInstallOptions.LocalServingHost,
		Port:    s.testEnv.WebhookInstallOptions.LocalServingPort,
	})
	s.NoError(err)

	s.reconciler = &DarkroomReconciler{
		Client: s.k8sClient,
		Log:    l.WithName("controllers").WithName("Darkroom"),
		Scheme: scheme,
	}

	s.NoError(s.reconciler.SetupControllerWithManager(s.mgr))
	s.NoError(s.reconciler.SetupWebhookWithManager(s.mgr))

	go func() {
		s.NoError(s.mgr.Start(s.stopCh))
	}()

	// wait for controller workers to start
	if !s.Eventually(func() bool {
		return strings.Contains(s.buf.String(), "Starting workers")
	}, 10*time.Second, 100*time.Millisecond) {
		s.Fail("unable to start workers via controller manager")
		os.Exit(1)
	}
}

func (s *DarkroomControllerSuite) SetupTest() {
	s.buf.Reset()
}

func (s *DarkroomControllerSuite) TestReconcile() {
	testcases := []struct {
		name             string
		ctx              context.Context
		darkroom         *deploymentsv1alpha1.Darkroom
		preReconcileRun  func(context.Context, client.Client, *deploymentsv1alpha1.Darkroom) error
		postReconcileRun func(context.Context, client.Client, *deploymentsv1alpha1.Darkroom) error
	}{
		{
			name: "Darkroom object status contains domain list",
			ctx:  context.Background(),
			darkroom: &deploymentsv1alpha1.Darkroom{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "darkroom-sample",
					Namespace: "default",
				},
				Spec: deploymentsv1alpha1.DarkroomSpec{
					Source: deploymentsv1alpha1.Source{
						Type: deploymentsv1alpha1.WebFolder,
						WebFolderMeta: deploymentsv1alpha1.WebFolderMeta{
							BaseURL: "https://example.com/assets/images",
						},
					},
					Domains: []string{"test.darkroom.net"},
				},
			},
			preReconcileRun: func(ctx context.Context, c client.Client, d *deploymentsv1alpha1.Darkroom) error {
				return c.Create(ctx, d)
			},
			postReconcileRun: func(ctx context.Context, c client.Client, d *deploymentsv1alpha1.Darkroom) error {
				desired := &deploymentsv1alpha1.Darkroom{}
				if err := c.Get(ctx, client.ObjectKey{Name: d.Name, Namespace: d.Namespace}, desired); err != nil {
					return err
				}
				s.Equal(d.Spec.Domains, desired.Status.Domains)
				return nil
			},
		},
		{
			name: "Darkroom object is not found",
			ctx:  context.Background(),
			darkroom: &deploymentsv1alpha1.Darkroom{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "darkroom-missing",
					Namespace: "default",
				},
			},
			preReconcileRun: func(ctx context.Context, c client.Client, d *deploymentsv1alpha1.Darkroom) error {
				// hack to run reconcile func and check if err is nil
				_, err := s.reconciler.Reconcile(ctrl.Request{
					NamespacedName: types.NamespacedName{
						Namespace: d.Namespace,
						Name:      d.Name,
					},
				})
				// should return nil if error was created with NewNotFound
				s.NoError(err)
				return nil
			},
			postReconcileRun: func(ctx context.Context, c client.Client, d *deploymentsv1alpha1.Darkroom) error {
				desired := &deploymentsv1alpha1.Darkroom{}
				err := c.Get(ctx, client.ObjectKey{Name: d.Name, Namespace: d.Namespace}, desired)
				s.Error(err)
				return nil
			},
		},
		{
			name: "Reconciler creates the required ConfigMap",
			ctx:  context.Background(),
			darkroom: &deploymentsv1alpha1.Darkroom{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "darkroom-config",
					Namespace: "default",
				},
				Spec: deploymentsv1alpha1.DarkroomSpec{
					Source: deploymentsv1alpha1.Source{
						Type: deploymentsv1alpha1.WebFolder,
						WebFolderMeta: deploymentsv1alpha1.WebFolderMeta{
							BaseURL: "https://example.com/assets/images",
						},
					},
					Domains: []string{"config.darkroom.net"},
				},
			},
			preReconcileRun: func(ctx context.Context, c client.Client, d *deploymentsv1alpha1.Darkroom) error {
				return c.Create(ctx, d)
			},
			postReconcileRun: func(ctx context.Context, c client.Client, d *deploymentsv1alpha1.Darkroom) error {
				cfgMap := &corev1.ConfigMap{}
				desiredMap := map[string]string{
					"CACHE_TIME":                            "31536000",
					"DEBUG":                                 "false",
					"LOG_LEVEL":                             "info",
					"PORT":                                  "3000",
					"SOURCE_BASEURL":                        "https://example.com/assets/images",
					"SOURCE_HYSTRIX_COMMANDNAME":            "WEBFOLDER_ADAPTER",
					"SOURCE_HYSTRIX_ERRORPERCENTTHRESHOLD":  "25",
					"SOURCE_HYSTRIX_MAXCONCURRENTREQUESTS":  "100",
					"SOURCE_HYSTRIX_REQUESTVOLUMETHRESHOLD": "10",
					"SOURCE_HYSTRIX_SLEEPWINDOW":            "10",
					"SOURCE_HYSTRIX_TIMEOUT":                "5000",
					"SOURCE_KIND":                           "WebFolder",
				}
				err := c.Get(ctx, client.ObjectKey{Name: d.Name, Namespace: d.Namespace}, cfgMap)
				s.NoError(err)
				s.Equal(desiredMap, cfgMap.Data)
				s.True(len(cfgMap.OwnerReferences) > 0)
				s.Equal(deploymentsv1alpha1.GroupVersion.String(), cfgMap.OwnerReferences[0].APIVersion)
				return nil
			},
		},
		{
			name: "Reconciler creates the required Deployment and updates state as deploying",
			ctx:  context.Background(),
			darkroom: &deploymentsv1alpha1.Darkroom{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "darkroom-deployment",
					Namespace: "default",
				},
				Spec: deploymentsv1alpha1.DarkroomSpec{
					Source: deploymentsv1alpha1.Source{
						Type: deploymentsv1alpha1.WebFolder,
						WebFolderMeta: deploymentsv1alpha1.WebFolderMeta{
							BaseURL: "https://example.com/assets/images",
						},
					},
					Domains: []string{"deployment.darkroom.net"},
				},
			},
			preReconcileRun: func(ctx context.Context, c client.Client, d *deploymentsv1alpha1.Darkroom) error {
				return c.Create(ctx, d)
			},
			postReconcileRun: func(ctx context.Context, c client.Client, d *deploymentsv1alpha1.Darkroom) error {
				depl := &appsv1.Deployment{}
				if err := c.Get(ctx, client.ObjectKey{Name: d.Name, Namespace: d.Namespace}, depl); err != nil {
					return err
				}
				desired := &deploymentsv1alpha1.Darkroom{}
				if err := c.Get(ctx, client.ObjectKey{Name: d.Name, Namespace: d.Namespace}, desired); err != nil {
					return err
				}
				s.Equal(&metav1.LabelSelector{
					MatchLabels: map[string]string{"darkroom": d.Name},
				}, depl.Spec.Selector)
				s.Equal(metav1.ObjectMeta{
					Labels: map[string]string{"darkroom": d.Name},
				}, depl.Spec.Template.ObjectMeta)
				s.True(len(depl.Spec.Template.Spec.Containers) > 0)
				s.Equal("latest", strings.Split(depl.Spec.Template.Spec.Containers[0].Image, ":")[1])
				s.Equal(d.Name, depl.Spec.Template.Spec.Containers[0].EnvFrom[0].ConfigMapRef.Name)
				s.Equal(deploymentsv1alpha1.Deploying, desired.Status.DeployState)
				s.True(len(depl.OwnerReferences) > 0)
				s.Equal(deploymentsv1alpha1.GroupVersion.String(), depl.OwnerReferences[0].APIVersion)
				return nil
			},
		},
	}

	for _, t := range testcases {
		s.SetupTest()
		s.Run(t.name, func() {
			s.NoError(t.preReconcileRun(t.ctx, s.k8sClient, t.darkroom))

			s.Eventually(func() bool {
				err := t.postReconcileRun(t.ctx, s.k8sClient, t.darkroom)
				s.NoError(err)
				return err == nil
			}, 2*time.Second, 100*time.Millisecond)
		})
	}
}

func (s *DarkroomControllerSuite) TearDownSuite() {
	close(s.stopCh)
	s.NoError(s.testEnv.Stop())
}
