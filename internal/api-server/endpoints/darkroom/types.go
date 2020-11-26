package darkroom

import "github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"

type Source struct {
	Type    string `json:"type"`
	BaseURL string `json:"baseUrl,omitempty"`
}

type Darkroom struct {
	Name        string               `json:"name"`
	Version     string               `json:"version"`
	Source      Source               `json:"source"`
	Domains     []string             `json:"domains,omitempty"`
	DeployState v1alpha1.DeployState `json:"deployState,omitempty"`
}

type List struct {
	Items []Darkroom `json:"items"`
}
