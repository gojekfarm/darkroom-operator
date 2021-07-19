package darkroom

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/mock"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
)

func (s *EndpointSuite) TestDelete() {

	s.Run("Success", func() {
		s.SetupTest()
		var co []client.DeleteOption
		s.mockClient.On("Get",
			mock.Anything,
			client.ObjectKey{
				Name:      "darkroom-get-sample",
				Namespace: "default",
			},
			mock.AnythingOfType("*v1alpha1.Darkroom"),
		).Return(nil)
		s.mockClient.On("Delete",
			mock.Anything,
			mock.AnythingOfType("*v1alpha1.Darkroom"),
			co,
		).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/default/darkrooms/darkroom-get-sample", nil)
		resp := httptest.NewRecorder()
		s.handler.ServeHTTP(resp, req)

		s.Equal("", resp.Body.String())
		s.Equal(http.StatusNoContent, resp.Code)

		s.mockClient.AssertExpectations(s.T())
	})

	s.Run("NotFound", func() {
		s.SetupTest()
		s.mockClient.On("Get",
			mock.Anything,
			client.ObjectKey{
				Name:      "darkroom-delete-sample",
				Namespace: "default",
			},
			mock.AnythingOfType("*v1alpha1.Darkroom"),
		).Return(apiErrors.NewNotFound(schema.GroupResource{
			Group:    v1alpha1.GroupVersion.Group,
			Resource: "Darkroom",
		}, "darkroom-delete-sample"))

		req := httptest.NewRequest(http.MethodDelete, "/default/darkrooms/darkroom-delete-sample", nil)
		resp := httptest.NewRecorder()
		s.handler.ServeHTTP(resp, req)

		s.Equal(`{
 "message": "Unable to delete instance darkroom-delete-sample",
 "error": "Darkroom.deployments.gojek.io \"darkroom-delete-sample\" not found"
}`, resp.Body.String())
		s.Equal(http.StatusNotFound, resp.Code)

		s.mockClient.AssertExpectations(s.T())
	})

	s.Run("DeleteError", func() {
		s.SetupTest()
		var co []client.DeleteOption
		s.mockClient.On("Get",
			mock.Anything,
			client.ObjectKey{
				Name:      "darkroom-delete-sample",
				Namespace: "default",
			},
			mock.AnythingOfType("*v1alpha1.Darkroom"),
		).Return(nil)
		s.mockClient.On("Delete",
			mock.Anything,
			mock.AnythingOfType("*v1alpha1.Darkroom"),
			co,
		).Return(errors.New("permission denied"))

		req := httptest.NewRequest(http.MethodDelete, "/default/darkrooms/darkroom-delete-sample", nil)
		resp := httptest.NewRecorder()
		s.handler.ServeHTTP(resp, req)

		s.Equal(`{
 "message": "Unable to delete instance darkroom-delete-sample",
 "error": "permission denied"
}`, resp.Body.String())
		s.Equal(http.StatusFailedDependency, resp.Code)

		s.mockClient.AssertExpectations(s.T())
	})
}
