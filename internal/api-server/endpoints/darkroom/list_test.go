package darkroom

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *EndpointSuite) TestList() {
	s.Run("Success", func() {
		s.SetupTest()
		s.mockClient.On("List",
			mock.Anything,
			mock.AnythingOfType("*v1alpha1.DarkroomList"),
			[]client.ListOption{
				client.InNamespace("default"),
			},
		).Return(nil)

		req := httptest.NewRequest(http.MethodGet, "/default/darkrooms", nil)
		resp := httptest.NewRecorder()
		s.handler.ServeHTTP(resp, req)

		s.Equal(`{
 "metadata": {},
 "items": null
}`, resp.Body.String())

		s.mockClient.AssertExpectations(s.T())
	})

	s.Run("ListNamespaceError", func() {
		s.SetupTest()
		s.mockClient.On("List",
			mock.Anything,
			mock.AnythingOfType("*v1alpha1.DarkroomList"),
			[]client.ListOption{
				client.InNamespace("internal-error"),
			},
		).Return(errors.New("internal error"))

		req := httptest.NewRequest(http.MethodGet, "/internal-error/darkrooms", nil)
		resp := httptest.NewRecorder()
		s.handler.ServeHTTP(resp, req)
		s.Equal(`{
 "message": "Unable to list darkrooms instances",
 "error": "internal error"
}`, resp.Body.String())
		s.Equal(http.StatusFailedDependency, resp.Code)

		s.mockClient.AssertExpectations(s.T())
	})
}
