package darkroom

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (e *Endpoint) get(request *restful.Request, response *restful.Response) {
	ns := request.PathParameter("namespace")
	n := request.PathParameter("name")
	e.respond(response, func() error {
		d := new(v1alpha1.Darkroom)
		if err := e.client.Get(request.Request.Context(), client.ObjectKey{Namespace: ns, Name: n}, d); err != nil {
			return err
		}
		return response.WriteAsJson(d)
	}, "Unable to get instance "+n)
}
