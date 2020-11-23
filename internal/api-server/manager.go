package api_server

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

type Options struct {
	Scheme    *runtime.Scheme
	Namespace string
	Port      int
}

type Manager interface {
	Start(stop <-chan struct{}) error
}

func NewManager(config *rest.Config, options Options) (Manager, error) { return nil, nil }
