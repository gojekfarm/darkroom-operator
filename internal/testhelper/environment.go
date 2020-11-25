package testhelper

import (
	"path/filepath"

	admissionv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func NewTestEnvironment() *envtest.Environment {
	return &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "config", "crd", "bases")},
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			MutatingWebhooks: getMutationWebhooks(),
		},
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
