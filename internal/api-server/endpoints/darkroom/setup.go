package darkroom

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Endpoint struct {
	client client.Client
}

func (e *Endpoint) SetupWithWS(ws *restful.WebService) {
	ws.Route(ws.GET("{namespace}/darkrooms").To(e.list).
		Param(ws.PathParameter("namespace", "namespace of darkroom instances").DataType("string")).
		Doc("List of Darkrooms").
		Returns(http.StatusOK, "OK", &List{}))

	ws.Route(ws.GET("{namespace}/darkrooms/{name}").To(e.get).
		Param(ws.PathParameter("namespace", "namespace of darkroom instances").DataType("string")).
		Param(ws.PathParameter("name", "identifier of darkroom instance").DataType("string")).
		Doc("Get Darkroom Instance").
		Returns(http.StatusOK, "OK", &Darkroom{}))

	ws.Route(ws.DELETE("{namespace}/darkrooms/{name}").To(e.delete).
		Param(ws.PathParameter("namespace", "namespace of darkroom instances").DataType("string")).
		Param(ws.PathParameter("name", "identifier of darkroom instance").DataType("string")).
		Doc("Delete Darkroom Instance").
		Returns(http.StatusNoContent, "NO CONTENT", &Darkroom{}))
}

func NewEndpoint(client client.Client) *Endpoint {
	return &Endpoint{client: client}
}
