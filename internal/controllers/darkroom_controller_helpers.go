package controllers

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	deploymentsv1alpha1 "github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
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

	err := ctrl.SetControllerReference(&darkroom, &cfg, r.Scheme)
	return cfg, err
}

func (r *DarkroomReconciler) desiredDeployment(darkroom deploymentsv1alpha1.Darkroom, configMap corev1.ConfigMap) (appsv1.Deployment, error) {
	depl := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.String(), Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      darkroom.Name,
			Namespace: darkroom.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"darkroom": darkroom.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"darkroom": darkroom.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "darkroom",
							Image: fmt.Sprintf("gojektech/darkroom:%s", darkroom.Spec.Version),
							EnvFrom: []corev1.EnvFromSource{
								{
									ConfigMapRef: &corev1.ConfigMapEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: configMap.Name,
										},
									},
								},
							},
							Ports: []corev1.ContainerPort{
								{ContainerPort: 3000, Name: "http", Protocol: "TCP"},
							},
						},
					},
				},
			},
		},
	}

	err := ctrl.SetControllerReference(&darkroom, &depl, r.Scheme)
	return depl, err
}

func (r *DarkroomReconciler) desiredService(darkroom deploymentsv1alpha1.Darkroom) (corev1.Service, error) {
	svc := corev1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      darkroom.Name,
			Namespace: darkroom.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 8080, Protocol: "TCP", TargetPort: intstr.FromString("http")},
			},
			Selector: map[string]string{"darkroom": darkroom.Name},
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	err := ctrl.SetControllerReference(&darkroom, &svc, r.Scheme)
	return svc, err
}
