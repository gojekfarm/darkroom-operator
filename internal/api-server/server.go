package apiserver

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"

	"github.com/gojekfarm/darkroom-operator/internal/api-server/rest"
	pkglog "github.com/gojekfarm/darkroom-operator/pkg/log"
)

var runLog = pkglog.Log.WithName("api-server").WithName("run")

type apiServer struct {
	server *http.Server
}

func (as *apiServer) Address() string {
	return as.server.Addr
}

func newApiServer(port int, allowedDomains []string, em rest.EndpointManager) *apiServer {
	container := restful.NewContainer()
	container.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	em.Setup(container)

	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{restful.HEADER_AccessControlAllowOrigin},
		AllowedDomains: allowedDomains,
		Container:      container,
	}
	container.Filter(cors.Filter)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: container.ServeMux,
	}
	return &apiServer{
		server: srv,
	}
}

func (as *apiServer) Start(stop <-chan struct{}) error {
	errChan := make(chan error)
	go func() {
		err := as.server.ListenAndServe()
		if err != nil {
			switch err {
			case http.ErrServerClosed:
				runLog.Info("shutting down api-server")
			default:
				runLog.Error(err, "could not start an HTTP Server")
				errChan <- err
			}
		}
	}()
	runLog.Info("starting api-server", "interface", "0.0.0.0", "port", strings.Split(as.Address(), ":")[1])
	select {
	case <-stop:
		runLog.Info("shutting down api-server")
		return as.server.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}
