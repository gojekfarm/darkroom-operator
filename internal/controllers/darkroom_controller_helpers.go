package controllers

import (
	"fmt"
	"strings"

	deploymentsv1alpha1 "github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *DarkroomReconciler) desiredConfigMap(darkroom deploymentsv1alpha1.Darkroom) (corev1.ConfigMap, error) {
	cfg := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      darkroom.Name,
			Namespace: darkroom.Namespace,
		},
		Data: map[string]string{
			"DEBUG":          "false",
			"LOG_LEVEL":      "info",
			"SOURCE_KIND":    string(darkroom.Spec.Source.Type),
			"SOURCE_BASEURL": darkroom.Spec.Source.BaseURL,
			"PORT":           "3000",
			"CACHE_TIME":     "31536000",
			"SOURCE_HYSTRIX_COMMANDNAME": strings.ToUpper(
				fmt.Sprintf("%s_ADAPTER", darkroom.Spec.Source.Type),
			),
			"SOURCE_HYSTRIX_TIMEOUT":                "5000",
			"SOURCE_HYSTRIX_MAXCONCURRENTREQUESTS":  "100",
			"SOURCE_HYSTRIX_REQUESTVOLUMETHRESHOLD": "10",
			"SOURCE_HYSTRIX_SLEEPWINDOW":            "10",
			"SOURCE_HYSTRIX_ERRORPERCENTTHRESHOLD":  "25",
		},
	}

	_ = ctrl.SetControllerReference(&darkroom, &cfg, r.Scheme)
	return cfg, nil
}
