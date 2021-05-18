package v1alpha1

type Bucket struct {
	// +kubebuilder:validation:MinLength=3
	Name string `json:"name"`

	CredentialsJson string `json:"credentialsJson,omitempty"`
	AccessKey       string `json:"accessKey,omitempty"`
	SecretKey       string `json:"secretKey,omitempty"`
}
