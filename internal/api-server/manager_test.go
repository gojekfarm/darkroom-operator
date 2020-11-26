package apiserver

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache/informertest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/gojekfarm/darkroom-operator/internal/controllers"
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
	s.testEnv = testhelper.NewTestEnvironment([]string{filepath.Join("..", "..", "config", "crd", "bases")})
	r := &controllers.DarkroomReconciler{
		Log:    s.testEnv.GetLogger().WithName("controllers").WithName("Darkroom"),
		Scheme: testhelper.Scheme,
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
	testcases := []struct {
		name                string
		assertCreateErr     bool
		assertStartErr      bool
		cacheAlreadyStarted bool
		serverPort          *int
		newMapperFunc       func(cfg *rest.Config, opts ...apiutil.DynamicRESTMapperOption) (meta.RESTMapper, error)
		newCacheFunc        cache.NewCacheFunc
		newClientFunc       func(config *rest.Config, options client.Options) (client.Client, error)
	}{
		{
			name:            "Success",
			assertCreateErr: false,
			newMapperFunc:   apiutil.NewDynamicRESTMapper,
			newCacheFunc:    cache.New,
			newClientFunc:   client.New,
		},
		{
			name:                "SuccessWithCacheAlreadyStarted",
			assertCreateErr:     false,
			cacheAlreadyStarted: true,
			newMapperFunc:       apiutil.NewDynamicRESTMapper,
			newCacheFunc:        cache.New,
			newClientFunc:       client.New,
		},
		{
			name:            "MapperError",
			assertCreateErr: true,
			newMapperFunc: func(cfg *rest.Config, opts ...apiutil.DynamicRESTMapperOption) (meta.RESTMapper, error) {
				return nil, errors.New("unable to create mapper")
			},
			newCacheFunc:  cache.New,
			newClientFunc: client.New,
		},
		{
			name:            "CacheError",
			assertCreateErr: true,
			newMapperFunc:   apiutil.NewDynamicRESTMapper,
			newCacheFunc: func(config *rest.Config, opts cache.Options) (cache.Cache, error) {
				return nil, errors.New("unable to create cache")
			},
			newClientFunc: client.New,
		},
		{
			name:           "CacheStartError",
			assertStartErr: true,
			newMapperFunc:  apiutil.NewDynamicRESTMapper,
			newCacheFunc: func(config *rest.Config, opts cache.Options) (cache.Cache, error) {
				return &informertest.FakeInformers{Error: errors.New("unable to start cache")}, nil
			},
			newClientFunc: client.New,
		},
		{
			name:            "ClientError",
			assertCreateErr: true,
			newMapperFunc:   apiutil.NewDynamicRESTMapper,
			newCacheFunc:    cache.New,
			newClientFunc: func(config *rest.Config, options client.Options) (client.Client, error) {
				return nil, errors.New("unable to create client")
			},
		},
		{
			name:           "BadServerPort",
			assertStartErr: true,
			newMapperFunc:  apiutil.NewDynamicRESTMapper,
			newCacheFunc:   cache.New,
			newClientFunc:  client.New,
			serverPort:     &badPort,
		},
	}

	for _, t := range testcases {
		s.Run(t.name, func() {
			stopCh := make(chan struct{})
			errCh := make(chan error)
			mf := NewManager(NewManagerFuncOptions{
				NewDynamicRESTMapper: t.newMapperFunc,
				NewCache:             t.newCacheFunc,
				NewClient:            t.newClientFunc,
			})
			sp := 9999
			if t.serverPort != nil {
				sp = *t.serverPort
			}

			m, err := mf(s.testEnv.GetConfig(), Options{
				Scheme:         runtime.NewScheme(),
				Port:           sp,
				AllowedDomains: []string{},
			})
			if t.cacheAlreadyStarted {
				m.(*manager).started = true
			}

			if t.assertCreateErr {
				defer close(errCh)
				defer close(stopCh)
				s.Error(err)
			} else {
				s.NoError(err)
				go func() {
					defer close(errCh)
					errCh <- m.Start(stopCh)
				}()
				if t.assertStartErr {
					defer close(stopCh)
					s.Error(<-errCh)
				} else {
					close(stopCh)
					s.NoError(<-errCh)
				}
			}
		})
	}
}
