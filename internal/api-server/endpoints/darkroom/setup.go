package darkroom

import (
	"github.com/emicklei/go-restful"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Endpoint struct {
	client client.Client
}

func (e *Endpoint) SetupWithWS(ws *restful.WebService) {
	ws.Route(ws.GET("darkrooms").To(e.list).
		Doc("List of Darkrooms").
		Returns(200, "OK", &List{}))
}

func NewEndpoint(client client.Client) *Endpoint {
	return &Endpoint{client: client}
}
