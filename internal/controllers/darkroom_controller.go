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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gojekfarm/darkroom-operator/internal/controllers/setup"

	deploymentsv1alpha1 "github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
)

// DarkroomReconciler reconciles a Darkroom object
type DarkroomReconciler struct {
	setup.Controller
	setup.Webhook
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=deployments.gojek.io,resources=darkrooms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=deployments.gojek.io,resources=darkrooms/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=configmaps;services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete

func (r *DarkroomReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("darkroom", req.NamespacedName)
	var darkroom deploymentsv1alpha1.Darkroom

	if err := r.Get(ctx, req.NamespacedName, &darkroom); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	cfg, _ := r.desiredConfigMap(darkroom)
	depl, _ := r.desiredDeployment(darkroom, cfg)
	svc, _ := r.desiredService(darkroom)

	applyOptions := []client.PatchOption{client.ForceOwnership, client.FieldOwner("darkroom-controller")}

	_ = r.Patch(ctx, &cfg, client.Apply, applyOptions...)
	_ = r.Patch(ctx, &depl, client.Apply, applyOptions...)
	_ = r.Patch(ctx, &svc, client.Apply, applyOptions...)

	darkroom.Status.Domains = darkroom.Spec.Domains
	darkroom.Status.DeployState = deploymentsv1alpha1.Deploying
	_ = r.Status().Update(ctx, &darkroom)
	return ctrl.Result{}, nil
}

func (r *DarkroomReconciler) SetupControllerWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&deploymentsv1alpha1.Darkroom{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func (r *DarkroomReconciler) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&deploymentsv1alpha1.Darkroom{}).
		Complete()
}
