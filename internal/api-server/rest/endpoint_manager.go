package rest

import (
	"os"

	"github.com/emicklei/go-restful/v3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gojekfarm/darkroom-operator/internal/api-server/endpoints/darkroom"
	"github.com/gojekfarm/darkroom-operator/internal/version"
)

type EndpointManager interface {
	Setup(c *restful.Container)
}

func NewEndpointManager(client client.Client) EndpointManager {
	return &endpointManager{
		client: client,
		endpoints: []Endpoint{
			darkroom.NewEndpoint(client),
		},
	}
}

type endpointManager struct {
	client    client.Client
	endpoints []Endpoint
}

func (em *endpointManager) Setup(c *restful.Container) {
	ws := new(restful.WebService)
	ws.
		Path("/api/").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	em.addVersionEndpoint(ws)

	for _, ep := range em.endpoints {
		ep.SetupWithWS(ws)
	}

	c.Add(ws)
}

func (em *endpointManager) addVersionEndpoint(ws *restful.WebService) {
	hostname, _ := os.Hostname()
	ws.Route(ws.GET("/version").To(func(req *restful.Request, resp *restful.Response) {
		response := VersionResponse{
			Hostname: hostname,
			Tagline:  version.Product,
			Version:  version.Build.Version,
		}
		_ = resp.WriteAsJson(response)
	}))
}
