package runtime

import (
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	deploymentsv1alpha1 "github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(deploymentsv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func Scheme() *runtime.Scheme {
	return scheme
}
