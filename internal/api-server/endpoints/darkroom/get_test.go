package darkroom

import (
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/mock"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
)

func (s *EndpointSuite) TestGet() {

	s.Run("Success", func() {
		s.SetupTest()
		s.mockClient.On("Get",
			mock.Anything,
			client.ObjectKey{
				Name:      "darkroom-get-sample",
				Namespace: "default",
			},
			mock.AnythingOfType("*v1alpha1.Darkroom"),
		).Return(nil)

		req := httptest.NewRequest(http.MethodGet, "/default/darkrooms/darkroom-get-sample", nil)
		resp := httptest.NewRecorder()
		s.handler.ServeHTTP(resp, req)

		s.Equal(`{
 "metadata": {
  "creationTimestamp": null
 },
 "spec": {
  "version": "",
  "source": {
   "type": ""
  },
  "domains": null
 },
 "status": {
  "deployState": ""
 }
}`, resp.Body.String())
		s.Equal(http.StatusOK, resp.Code)

		s.mockClient.AssertExpectations(s.T())
	})

	s.Run("NotFound", func() {
		s.SetupTest()
		s.mockClient.On("Get",
			mock.Anything,
			client.ObjectKey{
				Name:      "darkroom-get-sample",
				Namespace: "not-found",
			},
			mock.AnythingOfType("*v1alpha1.Darkroom"),
		).Return(apiErrors.NewNotFound(schema.GroupResource{
			Group:    v1alpha1.GroupVersion.Group,
			Resource: "Darkroom",
		}, "darkroom-get-sample"))

		req := httptest.NewRequest(http.MethodGet, "/not-found/darkrooms/darkroom-get-sample", nil)
		resp := httptest.NewRecorder()
		s.handler.ServeHTTP(resp, req)

		s.Equal(`{
 "message": "Unable to get instance darkroom-get-sample",
 "error": "Darkroom.deployments.gojek.io \"darkroom-get-sample\" not found"
}`, resp.Body.String())
		s.Equal(http.StatusNotFound, resp.Code)

		s.mockClient.AssertExpectations(s.T())
	})
}
