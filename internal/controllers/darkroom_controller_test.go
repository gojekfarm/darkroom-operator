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
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gojekfarm/darkroom-operator/internal/runtime"
	"github.com/gojekfarm/darkroom-operator/internal/testhelper"
	deploymentsv1alpha1 "github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

type DarkroomControllerSuite struct {
	suite.Suite
	testEnv    testhelper.Environment
	reconciler *DarkroomReconciler
}

func TestDarkroomControllerSuite(t *testing.T) {
	suite.Run(t, new(DarkroomControllerSuite))
}

func (s *DarkroomControllerSuite) SetupSuite() {
	s.testEnv = testhelper.NewTestEnvironment("..", "..")
	s.reconciler = &DarkroomReconciler{
		Log:    s.testEnv.GetLogger().WithName("controllers").WithName("Darkroom"),
		Scheme: runtime.Scheme(),
	}
	s.testEnv.Add(s.reconciler)
	s.NoError(s.testEnv.Start())

	var err error
	s.reconciler.Client, err = testhelper.NewClient(s.testEnv.GetConfig())
	s.NoError(err)
}

func (s *DarkroomControllerSuite) SetupTest() {
	s.testEnv.ResetLogs()
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
				_, err := s.reconciler.Reconcile(ctx, ctrl.Request{
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
		{
			name: "Reconciler creates the required Service",
			ctx:  context.Background(),
			darkroom: &deploymentsv1alpha1.Darkroom{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "darkroom-service",
					Namespace: "default",
				},
				Spec: deploymentsv1alpha1.DarkroomSpec{
					Source: deploymentsv1alpha1.Source{
						Type: deploymentsv1alpha1.WebFolder,
						WebFolderMeta: deploymentsv1alpha1.WebFolderMeta{
							BaseURL: "https://example.com/assets/images",
						},
					},
					Domains: []string{"service.darkroom.net"},
				},
			},
			preReconcileRun: func(ctx context.Context, c client.Client, d *deploymentsv1alpha1.Darkroom) error {
				return c.Create(ctx, d)
			},
			postReconcileRun: func(ctx context.Context, c client.Client, d *deploymentsv1alpha1.Darkroom) error {
				svc := &corev1.Service{}
				if err := c.Get(ctx, client.ObjectKey{Name: d.Name, Namespace: d.Namespace}, svc); err != nil {
					return err
				}
				desired := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      d.Name,
						Namespace: d.Namespace,
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{Name: "http", Port: 8080, Protocol: "TCP", TargetPort: intstr.FromString("http")},
						},
						Selector: map[string]string{"darkroom": d.Name},
						Type:     corev1.ServiceTypeClusterIP,
					},
				}
				s.Equal(desired.ObjectMeta.Name, svc.ObjectMeta.Name)
				s.Equal(desired.ObjectMeta.Namespace, svc.ObjectMeta.Namespace)
				s.NotEqual("", svc.Spec.ClusterIP)
				s.Equal(desired.Spec.Ports, svc.Spec.Ports)
				s.Equal(desired.Spec.Selector, svc.Spec.Selector)
				s.Equal(desired.Spec.Type, svc.Spec.Type)
				return nil
			},
		},
	}

	for _, t := range testcases {
		s.SetupTest()
		s.Run(t.name, func() {
			s.NoError(t.preReconcileRun(t.ctx, s.reconciler.Client, t.darkroom))

			s.Eventually(func() bool {
				err := t.postReconcileRun(t.ctx, s.reconciler.Client, t.darkroom)
				s.NoError(err)
				return err == nil
			}, 10*time.Second, 250*time.Millisecond)
		})
	}
}

func (s *DarkroomControllerSuite) TearDownSuite() {
	s.NoError(s.testEnv.Stop())
}
