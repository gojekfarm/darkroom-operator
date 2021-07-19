package darkroom

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
)

func (s *EndpointSuite) TestCreate() {
	obj := v1alpha1.Darkroom{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "darkroom-create-sample",
		},
		Spec: v1alpha1.DarkroomSpec{
			Source: v1alpha1.Source{
				Type: v1alpha1.WebFolder,
				WebFolderMeta: v1alpha1.WebFolderMeta{
					BaseURL: "https://example.com",
				},
			},
			Domains: []string{"darkroom-create-sample.example.com"},
		},
	}

	s.Run("Success", func() {
		s.SetupTest()
		s.mockClient.On("Create",
			mock.Anything,
			mock.AnythingOfType("*v1alpha1.Darkroom"),
			[]client.CreateOption{client.FieldOwner("api-server")},
		).Return(nil)

		b := &bytes.Buffer{}
		s.NoError(json.NewEncoder(b).Encode(obj))

		req := httptest.NewRequest(http.MethodPost, "/default/darkrooms", b)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "*/*")
		resp := httptest.NewRecorder()
		s.handler.ServeHTTP(resp, req)

		s.Equal(`{
 "metadata": {
  "name": "darkroom-create-sample",
  "namespace": "default",
  "creationTimestamp": null
 },
 "spec": {
  "version": "",
  "source": {
   "type": "WebFolder",
   "baseUrl": "https://example.com"
  },
  "domains": [
   "darkroom-create-sample.example.com"
  ]
 },
 "status": {
  "deployState": ""
 }
}`, resp.Body.String())
		s.Equal(http.StatusCreated, resp.Code)

		s.mockClient.AssertExpectations(s.T())
	})

	s.Run("CreateError", func() {
		s.SetupTest()
		s.mockClient.On("Create",
			mock.Anything,
			mock.AnythingOfType("*v1alpha1.Darkroom"),
			[]client.CreateOption{client.FieldOwner("api-server")},
		).Return(errors.New("internal error"))

		b := &bytes.Buffer{}
		s.NoError(json.NewEncoder(b).Encode(obj))

		req := httptest.NewRequest(http.MethodPost, "/default/darkrooms", b)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "*/*")
		resp := httptest.NewRecorder()
		s.handler.ServeHTTP(resp, req)

		s.Equal(`{
 "message": "Unable to create instance",
 "error": "internal error"
}`, resp.Body.String())
		s.Equal(http.StatusFailedDependency, resp.Code)

		s.mockClient.AssertExpectations(s.T())
	})

	s.Run("ValidationError", func() {
		s.SetupTest()

		invalidObj := obj.DeepCopy()
		invalidObj.Spec.Source.BaseURL = ""

		b := &bytes.Buffer{}
		s.NoError(json.NewEncoder(b).Encode(invalidObj))

		req := httptest.NewRequest(http.MethodPost, "/default/darkrooms", b)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "*/*")
		resp := httptest.NewRecorder()
		s.handler.ServeHTTP(resp, req)

		s.Equal(`{
 "message": "Unable to create instance",
 "error": "Darkroom.deployments.gojek.io \"darkroom-create-sample\" is invalid: spec.source.baseUrl: Invalid value: \"\": parse \"\": empty url"
}`, resp.Body.String())
		s.Equal(http.StatusUnprocessableEntity, resp.Code)

		s.mockClient.AssertExpectations(s.T())
	})

	s.Run("BadRequest", func() {
		s.SetupTest()

		req := httptest.NewRequest(http.MethodPost, "/default/darkrooms", strings.NewReader("abc"))
		resp := httptest.NewRecorder()
		s.handler.ServeHTTP(resp, req)

		s.Equal("415: Unsupported Media Type", resp.Body.String())
		s.Equal(http.StatusUnsupportedMediaType, resp.Code)

		s.mockClient.AssertExpectations(s.T())
	})

	s.Run("ReadEntityError", func() {
		s.SetupTest()

		req := httptest.NewRequest(http.MethodPost, "/default/darkrooms", strings.NewReader("abc"))
		req.Header.Add("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		s.handler.ServeHTTP(resp, req)

		s.Equal(`{
 "message": "Unable to create instance",
 "error": "invalid character 'a' looking for beginning of value"
}`, resp.Body.String())
		s.Equal(http.StatusUnprocessableEntity, resp.Code)

		s.mockClient.AssertExpectations(s.T())
	})
}
