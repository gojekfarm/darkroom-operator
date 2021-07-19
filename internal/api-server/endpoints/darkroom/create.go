package darkroom

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (e *Endpoint) create(request *restful.Request, response *restful.Response) {
	e.respond(response, func() error {
		d := new(v1alpha1.Darkroom)
		if err := request.ReadEntity(d); err != nil {
			return err
		}
		if err := d.ValidateCreate(); err != nil {
			return err
		}
		d.Namespace = request.PathParameter("namespace")
		if err := e.client.Create(request.Request.Context(), d, &client.CreateOptions{FieldManager: "api-server"});
			err != nil {
			return err
		}
		return response.WriteHeaderAndEntity(http.StatusCreated, d)
	}, "Unable to create instance")
}
