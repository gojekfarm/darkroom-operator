package log

import kubelog "sigs.k8s.io/controller-runtime/pkg/log"

var (
	Log = kubelog.Log.WithName("darkroom-operator")
)
