package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"

	"github.com/gojekfarm/darkroom-operator/internal/runtime"
	"github.com/gojekfarm/darkroom-operator/internal/testhelper/mocks"
)

func TestVersionRouteOnNewEndpointManager(t *testing.T) {
	mc := &mocks.MockClient{RuntimeScheme: runtime.Scheme()}
	c := restful.NewContainer()

	em := NewEndpointManager(mc)
	em.Setup(c)

	req, _ := http.NewRequest(http.MethodGet, "/api/version", nil)
	resp := httptest.NewRecorder()

	c.ServeMux.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}
