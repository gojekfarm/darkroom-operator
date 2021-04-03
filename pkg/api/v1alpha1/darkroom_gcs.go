package v1alpha1

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	GoogleCloudStorage Type = "GoogleCloudStorage"
)

func (d *Darkroom) validateGoogleCloudStorage() *field.Error {
	if d.Spec.Source.Bucket == nil {
		return field.Required(
			field.NewPath("spec").Child("source").Child("bucket"),
			fmt.Sprintf("field required with Type %s", d.Spec.Source.Type),
		)
	}
	b := d.Spec.Source.Bucket
	if b.CredentialsJson == "" {
		return field.Required(
			field.NewPath("spec").Child("source").Child("bucket").Child("credentialsJson"),
			fmt.Sprintf("field required with Type %s", d.Spec.Source.Type),
		)
	}
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(b.CredentialsJson), &js); err != nil {
		return field.Invalid(
			field.NewPath("spec").Child("source").Child("bucket").Child("credentialsJson"),
			b.CredentialsJson,
			err.Error(),
		)
	}
	return nil
}
