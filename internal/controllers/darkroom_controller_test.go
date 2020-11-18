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
	"fmt"
	"path/filepath"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

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

	s.testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "config", "crd", "bases")},
	}
	s.stopCh = make(chan struct{})

	var err error
	s.cfg, err = s.testEnv.Start()
	s.NoError(err)
	s.NotNil(s.cfg)

	err = deploymentsv1alpha1.AddToScheme(scheme.Scheme)
	s.NoError(err)

	// +kubebuilder:scaffold:scheme

	s.k8sClient, err = client.New(s.cfg, client.Options{Scheme: scheme.Scheme})
	s.NoError(err)
	s.NotNil(s.k8sClient)

	s.mgr, err = ctrl.NewManager(s.cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	s.NoError(err)

	s.reconciler = &DarkroomReconciler{
		Client: s.k8sClient,
		Log:    l.WithName("controllers").WithName("Darkroom"),
		Scheme: scheme.Scheme,
	}

	s.NoError(s.reconciler.SetupWithManager(s.mgr))
	go func() {
		s.NoError(s.mgr.Start(s.stopCh))
	}()
}

func (s *DarkroomControllerSuite) SetupTest() {
	//s.buf.Reset()
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
				Spec: deploymentsv1alpha1.DarkroomSpec{
					Source: deploymentsv1alpha1.Source{
						Type: deploymentsv1alpha1.WebFolder,
						WebFolderMeta: deploymentsv1alpha1.WebFolderMeta{
							BaseURL: "https://example.com/assets/images",
						},
					},
					Domains: []string{"missing.darkroom.net"},
				},
			},
			preReconcileRun: func(ctx context.Context, c client.Client, d *deploymentsv1alpha1.Darkroom) error {
				if err := c.Create(ctx, d); err != nil {
					return err
				}
				// Delete as soon as created, but the Reconcile should be triggered
				return c.Delete(ctx, d)
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
	}

	for _, t := range testcases {
		s.SetupTest()
		s.Run(t.name, func() {
			s.NoError(t.preReconcileRun(t.ctx, s.k8sClient, t.darkroom))

			s.Eventually(func() bool {
				err := t.postReconcileRun(t.ctx, s.k8sClient, t.darkroom)
				s.NoError(err)
				fmt.Printf("============= logs: \n%s\n\n", s.buf.String())
				return err == nil
			}, 5*time.Second, 100*time.Millisecond)
		})
	}
}

func (s *DarkroomControllerSuite) TearDownSuite() {
	close(s.stopCh)
	s.NoError(s.testEnv.Stop())
}
