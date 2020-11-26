package setup

import ctrl "sigs.k8s.io/controller-runtime"

type Controller interface {
	SetupControllerWithManager(mgr ctrl.Manager) error
}
