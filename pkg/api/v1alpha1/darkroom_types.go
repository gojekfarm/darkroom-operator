/*
MIT License

Copyright (c) 2020 GO-JEK Tech

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:validation:Enum=WebFolder;S3;GoogleCloudStorage
type Type string

type DeployState string

const (
	Deploying DeployState = "Deploying"
)

type Source struct {
	// Type specifies storage backend to use with darkroom.
	// Valid values are:
	// - "WebFolder": simple storage backend to serve images from a hosted image source;
	// - "S3": storage backend to serve images from S3 backend;
	// - "GoogleCloudStorage": storage backend to serve images from GoogleCloudStorage backend;
	Type Type `json:"type"`

	WebFolderMeta `json:",inline"`

	Bucket *Bucket `json:"bucket,omitempty"`

	// +kubebuilder:default="/"
	// +optional
	Prefix string `json:"prefix,omitempty"`
}

// DarkroomSpec defines the desired state of Darkroom
type DarkroomSpec struct {
	// +optional
	Version string `json:"version"`

	Source Source `json:"source"`

	// +optional
	// PathPrefix prepends the prefix in the URL when serving images
	PathPrefix string `json:"pathPrefix,omitempty"`

	// +kubebuilder:validation:MinItems=1
	Domains []string `json:"domains"`
}

// DarkroomStatus defines the observed state of Darkroom
type DarkroomStatus struct {
	DeployState DeployState `json:"deployState"`
	// +optional
	Domains []string `json:"domains,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Darkroom is the Schema for the darkrooms API
type Darkroom struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DarkroomSpec   `json:"spec,omitempty"`
	Status DarkroomStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DarkroomList contains a list of Darkroom
type DarkroomList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Darkroom `json:"items"`
}

// +kubebuilder:webhook:path=/mutate-deployments-gojek-io-v1alpha1-darkroom,mutating=true,failurePolicy=fail,groups=deployments.gojek.io,resources=darkrooms,verbs=create;update,versions=v1alpha1,name=mdarkroom.gojek.io,sideEffects=None,admissionReviewVersions=v1;v1beta1
// +kubebuilder:webhook:path=/validate-deployments-gojek-io-v1alpha1-darkroom,mutating=false,failurePolicy=fail,groups=deployments.gojek.io,resources=darkrooms,verbs=create;update;delete,versions=v1alpha1,name=vdarkroom.gojek.io,sideEffects=None,admissionReviewVersions=v1;v1beta1

var _ webhook.Defaulter = &Darkroom{}
var _ webhook.Validator = &Darkroom{}

func init() {
	SchemeBuilder.Register(&Darkroom{}, &DarkroomList{})
}
