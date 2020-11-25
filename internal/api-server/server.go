package apiserver

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pkglog "github.com/gojekfarm/darkroom-operator/pkg/log"
)

var runLog = pkglog.Log.WithName("api-server").WithName("run")

type apiServer struct {
	server *http.Server
}

func (as *apiServer) Address() string {
	return as.server.Addr
}

func newApiServer(port int, allowedDomains []string, _ client.Client) *apiServer {
	container := restful.NewContainer()
	container.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: container.ServeMux,
	}

	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{restful.HEADER_AccessControlAllowOrigin},
		AllowedDomains: allowedDomains,
		Container:      container,
	}

	ws := new(restful.WebService)
	ws.
		Path("/").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	container.Add(ws)
	container.Filter(cors.Filter)
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
