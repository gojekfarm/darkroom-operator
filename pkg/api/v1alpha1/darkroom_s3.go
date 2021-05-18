package v1alpha1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	S3 Type = "S3"
)

func (d *Darkroom) validateS3() *field.Error {
	if d.Spec.Source.Bucket == nil {
		return field.Required(
			field.NewPath("spec").Child("source").Child("bucket"),
			fmt.Sprintf("field required with Type %s", d.Spec.Source.Type),
		)
	}
	b := d.Spec.Source.Bucket
	if b.AccessKey == "" {
		return field.Required(
			field.NewPath("spec").Child("source").Child("bucket").Child("accessKey"),
			fmt.Sprintf("field required with Type %s", d.Spec.Source.Type),
		)
	}
	if b.SecretKey == "" {
		return field.Required(
			field.NewPath("spec").Child("source").Child("bucket").Child("secretKey"),
			fmt.Sprintf("field required with Type %s", d.Spec.Source.Type),
		)
	}
	return nil
}
