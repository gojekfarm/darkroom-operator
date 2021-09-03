package darkroom

import (
	"encoding/json"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
)

func (e *Endpoint) respond(response *restful.Response, f func() error, errMsg string) {
	if err := f(); err != nil {
		code := http.StatusFailedDependency
		errStr := err.Error()

		switch err := err.(type) {
		case *apiErrors.StatusError:
			code = int(err.Status().Code)
			errStr = err.Status().Message
		case restful.ServiceError:
			code = err.Code
			errMsg = err.Message
		case *json.SyntaxError:
			code = http.StatusUnprocessableEntity
		}

		response.WriteHeader(code)
		_ = response.WriteAsJson(Error{
			Message: errMsg,
			Err:     errStr,
		})
	}
}
