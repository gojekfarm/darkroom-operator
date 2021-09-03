package rest

import "github.com/emicklei/go-restful/v3"

type Endpoint interface {
	SetupWithWS(ws *restful.WebService)
}

type VersionResponse struct {
	Hostname string `json:"hostname"`
	Tagline  string `json:"tagline"`
	Version  string `json:"version"`
}
