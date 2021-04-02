package apiserver

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	mgr "sigs.k8s.io/controller-runtime/pkg/manager"

	endpointRest "github.com/gojekfarm/darkroom-operator/internal/api-server/rest"
)

var (
	defaultRetryPeriod = 2 * time.Second
)

type NewManagerFunc func(config *rest.Config, options Options) (Manager, error)

type NewManagerFuncOptions struct {
	NewDynamicRESTMapper func(cfg *rest.Config, opts ...apiutil.DynamicRESTMapperOption) (meta.RESTMapper, error)
	NewCache             cache.NewCacheFunc
	NewClientBuilder     mgr.ClientBuilder
}

type Options struct {
	Scheme                *runtime.Scheme
	Namespace             string
	Port                  int
	ClientBuilder         mgr.ClientBuilder
	ClientDisableCacheFor []client.Object
	AllowedDomains        []string
}

type Manager interface {
	Start(context.Context) error
}

type manager struct {
	em              endpointRest.EndpointManager
	started         bool
	internalCtx     context.Context
	internalCancel  context.CancelFunc
	internalStop    <-chan struct{}
	internalStopper chan<- struct{}
	cache           cache.Cache
	errChan         chan error
	port            int
	allowedDomains  []string
}

func NewManager(newOpts NewManagerFuncOptions) NewManagerFunc {
	return func(config *rest.Config, options Options) (Manager, error) {
		mapper, err := newOpts.NewDynamicRESTMapper(config)
		if err != nil {
			return nil, err
		}

		cc, err := newOpts.NewCache(config, cache.Options{
			Scheme:    options.Scheme,
			Mapper:    mapper,
			Resync:    &defaultRetryPeriod,
			Namespace: options.Namespace,
		})
		if err != nil {
			return nil, err
		}

		c, err := newOpts.NewClientBuilder.
			WithUncached(options.ClientDisableCacheFor...).
			Build(cc, config, client.Options{Scheme: options.Scheme, Mapper: mapper})
		if err != nil {
			return nil, err
		}

		em := endpointRest.NewEndpointManager(c)

		stop := make(chan struct{})
		return &manager{
			cache:           cc,
			em:              em,
			internalStop:    stop,
			internalStopper: stop,
			errChan:         make(chan error),
			port:            options.Port,
			allowedDomains:  options.AllowedDomains,
		}, nil
	}
}

func (m *manager) Start(ctx context.Context) error {
	m.internalCtx, m.internalCancel = context.WithCancel(ctx)

	m.waitForCache()
	srv := newApiServer(m.port, m.allowedDomains, m.em)

	go func() {
		if err := srv.Start(m.internalStop); err != nil {
			m.errChan <- err
		}
	}()
	select {
	case <-ctx.Done():
		return nil
	case err := <-m.errChan:
		return err
	}
}

func (m *manager) waitForCache() {
	if m.started {
		return
	}

	go func() {
		if err := m.cache.Start(m.internalCtx); err != nil {
			m.errChan <- err
		}
	}()

	m.cache.WaitForCacheSync(m.internalCtx)
	m.started = true
}
