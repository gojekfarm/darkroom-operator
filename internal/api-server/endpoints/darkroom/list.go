package darkroom

import (
	"github.com/emicklei/go-restful/v3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
)

func (e *Endpoint) list(request *restful.Request, response *restful.Response) {
	ns := request.PathParameter("namespace")
	e.respond(response, func() error {
		dl := new(v1alpha1.DarkroomList)
		if err := e.client.List(request.Request.Context(), dl, client.InNamespace(ns)); err != nil {
			return err
		}
		return response.WriteAsJson(dl)
	}, "Unable to list darkrooms instances")
}
