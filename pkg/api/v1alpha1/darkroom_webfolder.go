package v1alpha1

import (
	"net/url"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	WebFolder Type = "WebFolder"
)

type WebFolderMeta struct {
	BaseURL string `json:"baseUrl,omitempty"`
}

func (d *Darkroom) validateWebFolder() *field.Error {
	if _, err := url.ParseRequestURI(d.Spec.Source.BaseURL); err != nil {
		return field.Invalid(
			field.NewPath("spec").Child("source").Child("baseUrl"),
			d.Spec.Source.BaseURL,
			err.Error(),
		)
	}
	return nil
}
