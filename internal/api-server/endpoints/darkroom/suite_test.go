package darkroom

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gojekfarm/darkroom-operator/internal/controllers"
	"github.com/gojekfarm/darkroom-operator/internal/testhelper"
	"github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
)

type EndpointSuite struct {
	suite.Suite
	ep         *Endpoint
	testEnv    testhelper.Environment
	container  *restful.Container
	reconciler *controllers.DarkroomReconciler
}

func TestEndpoint(t *testing.T) {
	suite.Run(t, new(EndpointSuite))
}

func (s *EndpointSuite) SetupSuite() {
	s.testEnv = testhelper.NewTestEnvironment([]string{filepath.Join("..", "..", "..", "..", "config", "crd", "bases")})
	s.reconciler = &controllers.DarkroomReconciler{
		Log:    s.testEnv.GetLogger().WithName("controllers").WithName("Darkroom"),
		Scheme: testhelper.Scheme,
	}
	s.testEnv.Add(s.reconciler)

	s.NoError(s.testEnv.Start())

	c, err := testhelper.NewClient(s.testEnv.GetConfig())
	s.NoError(err)
	s.reconciler.Client = c

	s.ep = NewEndpoint(c)
	ws := new(restful.WebService)
	s.ep.SetupWithWS(ws)
	s.container = restful.NewContainer()
	s.container.Add(ws)
}

func (s *EndpointSuite) TearDownSuite() {
	s.NoError(s.testEnv.Stop())
}

func (s *EndpointSuite) TestList() {
	ctx := context.Background()

	d := &v1alpha1.Darkroom{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "darkroom-list-sample",
			Namespace: "default",
		},
		Spec: v1alpha1.DarkroomSpec{
			Source: v1alpha1.Source{
				Type: v1alpha1.WebFolder,
				WebFolderMeta: v1alpha1.WebFolderMeta{
					BaseURL: "https://example.com",
				},
			},
			Domains: []string{"darkroom-list-sample.example.com"},
		},
	}
	s.NoError(s.ep.client.Create(ctx, d))

	s.Eventually(func() bool {
		req := httptest.NewRequest(http.MethodGet, "/darkrooms", nil)
		resp := httptest.NewRecorder()
		s.container.ServeHTTP(resp, req)
		return resp.Body.String() == `{
 "items": [
  {
   "name": "darkroom-list-sample",
   "version": "latest",
   "source": {
    "type": "WebFolder",
    "baseUrl": "https://example.com"
   },
   "domains": [
    "darkroom-list-sample.example.com"
   ],
   "deployState": "Deploying"
  }
 ]
}`
	}, 2*time.Second, 100*time.Millisecond)
}
