package apiserver

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache/informertest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	mgr "sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/gojekfarm/darkroom-operator/internal/controllers"
	"github.com/gojekfarm/darkroom-operator/internal/runtime"
	"github.com/gojekfarm/darkroom-operator/internal/testhelper"
)

type ManagerSuite struct {
	suite.Suite
	testEnv testhelper.Environment
}

func TestManager(t *testing.T) {
	suite.Run(t, new(ManagerSuite))
}

func (s *ManagerSuite) SetupSuite() {
	s.testEnv = testhelper.NewTestEnvironment("..", "..")
	r := &controllers.DarkroomReconciler{
		Log:    s.testEnv.GetLogger().WithName("controllers").WithName("Darkroom"),
		Scheme: runtime.Scheme(),
	}
	s.testEnv.Add(r)

	s.NoError(s.testEnv.Start())
	c, _ := testhelper.NewClient(s.testEnv.GetConfig())
	r.Client = c
}

func (s *ManagerSuite) TearDownSuite() {
	s.NoError(s.testEnv.Stop())
}

func (s *ManagerSuite) TestStart() {
	badPort := -9999
	clientBuilder := mgr.NewClientBuilder()
	testcases := []struct {
		name                string
		assertCreateErr     bool
		assertStartErr      bool
		cacheAlreadyStarted bool
		serverPort          *int
		newMapperFunc       func(cfg *rest.Config, opts ...apiutil.DynamicRESTMapperOption) (meta.RESTMapper, error)
		newCacheFunc        cache.NewCacheFunc
		newClientBuilder    mgr.ClientBuilder
	}{
		{
			name:             "Success",
			assertCreateErr:  false,
			newMapperFunc:    apiutil.NewDynamicRESTMapper,
			newCacheFunc:     cache.New,
			newClientBuilder: clientBuilder,
		},
		{
			name:                "SuccessWithCacheAlreadyStarted",
			assertCreateErr:     false,
			cacheAlreadyStarted: true,
			newMapperFunc:       apiutil.NewDynamicRESTMapper,
			newCacheFunc:        cache.New,
			newClientBuilder:    clientBuilder,
		},
		{
			name:            "MapperError",
			assertCreateErr: true,
			newMapperFunc: func(cfg *rest.Config, opts ...apiutil.DynamicRESTMapperOption) (meta.RESTMapper, error) {
				return nil, errors.New("unable to create mapper")
			},
			newCacheFunc:     cache.New,
			newClientBuilder: clientBuilder,
		},
		{
			name:            "CacheError",
			assertCreateErr: true,
			newMapperFunc:   apiutil.NewDynamicRESTMapper,
			newCacheFunc: func(config *rest.Config, opts cache.Options) (cache.Cache, error) {
				return nil, errors.New("unable to create cache")
			},
			newClientBuilder: clientBuilder,
		},
		{
			name:           "CacheStartError",
			assertStartErr: true,
			newMapperFunc:  apiutil.NewDynamicRESTMapper,
			newCacheFunc: func(config *rest.Config, opts cache.Options) (cache.Cache, error) {
				return &informertest.FakeInformers{Error: errors.New("unable to start cache")}, nil
			},
			newClientBuilder: clientBuilder,
		},
		{
			name:             "ClientError",
			assertCreateErr:  true,
			newMapperFunc:    apiutil.NewDynamicRESTMapper,
			newCacheFunc:     cache.New,
			newClientBuilder: &badClientBuilder{},
		},
		{
			name:             "BadServerPort",
			assertStartErr:   true,
			newMapperFunc:    apiutil.NewDynamicRESTMapper,
			newCacheFunc:     cache.New,
			newClientBuilder: clientBuilder,
			serverPort:       &badPort,
		},
	}

	for _, t := range testcases {
		s.Run(t.name, func() {
			ctx, cancel := context.WithCancel(context.Background())
			errCh := make(chan error)
			mf := NewManager(NewManagerFuncOptions{
				NewDynamicRESTMapper: t.newMapperFunc,
				NewCache:             t.newCacheFunc,
				NewClientBuilder:     t.newClientBuilder,
			})
			sp := 9999
			if t.serverPort != nil {
				sp = *t.serverPort
			}

			m, err := mf(s.testEnv.GetConfig(), Options{
				Scheme:         runtime.Scheme(),
				Port:           sp,
				AllowedDomains: []string{},
			})
			if t.cacheAlreadyStarted {
				m.(*manager).started = true
			}

			if t.assertCreateErr {
				defer close(errCh)
				defer cancel()
				s.Error(err)
			} else {
				s.NoError(err)
				go func() {
					defer close(errCh)
					errCh <- m.Start(ctx)
				}()
				if t.assertStartErr {
					defer cancel()
					s.Error(<-errCh)
				} else {
					cancel()
					s.NoError(<-errCh)
				}
			}
		})
	}
}

type badClientBuilder struct {
}

func (b *badClientBuilder) WithUncached(objs ...client.Object) mgr.ClientBuilder {
	return b
}

func (b *badClientBuilder) Build(cache cache.Cache, config *rest.Config, options client.Options) (client.Client, error) {
	return nil, errors.New("unable to create client")
}
