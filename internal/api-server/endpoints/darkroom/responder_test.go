package darkroom

import (
	"net/http"
	"net/http/httptest"

	"github.com/emicklei/go-restful/v3"

	"github.com/gojekfarm/darkroom-operator/internal/testhelper/mocks"
)

func (s *EndpointSuite) TestResponder() {
	e := NewEndpoint(mocks.NewMockClient(s.T()))

	e.respond(restful.NewResponse(httptest.NewRecorder()), func() error {
		return restful.NewError(http.StatusUnprocessableEntity, "unprocessable")
	}, "err message")
}
