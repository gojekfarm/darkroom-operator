package darkroom

import (
	"net/http"

	"github.com/emicklei/go-restful/v3"
)

func (e *Endpoint) respond(response *restful.Response, f func() error, errMsg string) {
	if err := f(); err != nil {
		_ = response.WriteAsJson(Error{
			Message: errMsg,
			Err:     err.Error(),
			Code:    http.StatusInternalServerError,
		})
	}
}
