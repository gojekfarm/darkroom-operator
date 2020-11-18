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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	deploymentsv1alpha1 "github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
)

// DarkroomReconciler reconciles a Darkroom object
type DarkroomReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=deployments.gojek.io,resources=darkrooms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=deployments.gojek.io,resources=darkrooms/status,verbs=get;update;patch

func (r *DarkroomReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	l := r.Log.WithValues("darkroom", req.NamespacedName)
	var darkroom deploymentsv1alpha1.Darkroom

	if err := r.Get(ctx, req.NamespacedName, &darkroom); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	cfg, err := r.desiredConfigMap(darkroom)
	if err != nil {
		l.Error(err, "cfg err")
		return ctrl.Result{}, err
	}

	applyOptions := []client.PatchOption{client.ForceOwnership, client.FieldOwner("darkroom-controller")}

	if err := r.Patch(ctx, &cfg, client.Apply, applyOptions...); err != nil {
		l.Error(err, "patch err")
		return ctrl.Result{}, err
	}
	l.Info("patched cgfMap")

	darkroom.Status.Domains = darkroom.Spec.Domains
	if err := r.Status().Update(ctx, &darkroom); err != nil {
		l.Error(err, "status err")
		return ctrl.Result{}, err
	}
	l.Info("updated status object")
	return ctrl.Result{}, nil
}

func (r *DarkroomReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&deploymentsv1alpha1.Darkroom{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
