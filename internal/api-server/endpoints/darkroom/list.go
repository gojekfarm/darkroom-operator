package darkroom

import (
	"github.com/emicklei/go-restful"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
)

func (e *Endpoint) list(request *restful.Request, response *restful.Response) {
	dl := new(v1alpha1.DarkroomList)
	_ = e.client.List(request.Request.Context(), dl, &client.ListOptions{})
	l := From.List(dl)
	_ = response.WriteAsJson(l)
}
