package setup

import ctrl "sigs.k8s.io/controller-runtime"

type Webhook interface {
	SetupWebhookWithManager(mgr ctrl.Manager) error
}
