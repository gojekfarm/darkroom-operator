package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func (d *Darkroom) Default() {
	log.Info("default", "name", d.Name)

	if d.Spec.Version == "" {
		d.Spec.Version = "latest"
	}
}

func (d *Darkroom) ValidateCreate() error {
	log.Info("validate create", "name", d.Name)
	var allErrs field.ErrorList
	switch d.Spec.Source.Type {
	case WebFolder:
		if err := d.validateWebFolder(); err != nil {
			allErrs = append(allErrs, err)
		}
	case S3:
		if err := d.validateS3(); err != nil {
			allErrs = append(allErrs, err)
		}
	case GoogleCloudStorage:
		if err := d.validateGoogleCloudStorage(); err != nil {
			allErrs = append(allErrs, err)
		}
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(
		schema.GroupKind{Group: GroupVersion.Group, Kind: "Darkroom"},
		d.Name, allErrs)
}

func (d *Darkroom) ValidateUpdate(old runtime.Object) error {
	log.Info("validate update", "name", d.Name)
	return nil
}

func (d *Darkroom) ValidateDelete() error {
	log.Info("validate delete", "name", d.Name)
	return nil
}
