package darkroom

import (
	"testing"

	"github.com/emicklei/go-restful/v3"
	"github.com/stretchr/testify/suite"

	"github.com/gojekfarm/darkroom-operator/internal/testhelper/mocks"
)

type EndpointSuite struct {
	suite.Suite
	handler    *restful.Container
	mockClient *mocks.MockClient
}

func TestEndpoint(t *testing.T) {
	suite.Run(t, new(EndpointSuite))
}

func (s *EndpointSuite) SetupTest() {
	s.mockClient = mocks.NewMockClient(s.T())
	ws := new(restful.WebService)
	ws.Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	NewEndpoint(s.mockClient).SetupWithWS(ws)
	s.handler = restful.NewContainer()
	s.handler.Add(ws)
}
