package darkroom

import (
	"errors"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
)

func (e *Endpoint) respond(response *restful.Response, f func() error, errMsg string) {
	if err := f(); err != nil {
		apiErr := &apiErrors.StatusError{}
		if errors.As(err, &apiErr) {
			_ = response.WriteAsJson(Error{
				Message: errMsg,
				Err:     apiErr.Status().Message,
				Code:    int(apiErr.Status().Code),
			})
			return
		}
		_ = response.WriteAsJson(Error{
			Message: errMsg,
			Err:     err.Error(),
			Code:    http.StatusFailedDependency,
		})
	}
}
