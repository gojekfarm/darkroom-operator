package testhelper

import (
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gojekfarm/darkroom-operator/internal/runtime"
)

func NewClient(config *rest.Config) (client.Client, error) {
	return client.New(config, client.Options{Scheme: runtime.Scheme()})
}
