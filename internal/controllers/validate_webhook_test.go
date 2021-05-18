package controllers

import (
	"context"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	deploymentsv1alpha1 "github.com/gojekfarm/darkroom-operator/pkg/api/v1alpha1"
)

func (s *DarkroomControllerSuite) TestCreate() {
	testcases := []struct {
		name string
		obj  client.Object
	}{
		{
			name: "WebFolder",
			obj: &deploymentsv1alpha1.Darkroom{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: deploymentsv1alpha1.DarkroomSpec{
					Source: deploymentsv1alpha1.Source{
						Type:          deploymentsv1alpha1.WebFolder,
						WebFolderMeta: deploymentsv1alpha1.WebFolderMeta{BaseURL: "https://example.com"},
					},
					Domains: []string{"test.darkroom.com"},
				},
			},
		},
		{
			name: "S3",
			obj: &deploymentsv1alpha1.Darkroom{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: deploymentsv1alpha1.DarkroomSpec{
					Source: deploymentsv1alpha1.Source{
						Type: deploymentsv1alpha1.S3,
						Bucket: &deploymentsv1alpha1.Bucket{
							Name:      "abc-test",
							AccessKey: "some-key",
							SecretKey: "super-secret",
						},
					},
					Domains: []string{"test.darkroom.com"},
				},
			},
		},
		{
			name: "GoogleCloudStorage",
			obj: &deploymentsv1alpha1.Darkroom{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: deploymentsv1alpha1.DarkroomSpec{
					Source: deploymentsv1alpha1.Source{
						Type: deploymentsv1alpha1.GoogleCloudStorage,
						Bucket: &deploymentsv1alpha1.Bucket{
							Name:            "abc-test",
							CredentialsJson: `{}`,
						},
					},
					Domains: []string{"test.darkroom.com"},
				},
			},
		},
	}

	for _, tc := range testcases {
		s.SetupTest()
		s.Run(tc.name, func() {
			err := s.client.Create(context.Background(), tc.obj, client.DryRunAll)
			s.NoError(err)
		})
	}
}

func (s *DarkroomControllerSuite) TestCreateError() {
	testcases := []struct {
		name      string
		obj       client.Object
		errString string
	}{
		{
			name: "FailWebFolderWhenBaseUrlIsMissing",
			obj: &deploymentsv1alpha1.Darkroom{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: deploymentsv1alpha1.DarkroomSpec{
					Source: deploymentsv1alpha1.Source{
						Type: deploymentsv1alpha1.WebFolder,
					},
					Domains: []string{"test.darkroom.com"},
				},
			},
			errString: `spec.source.baseUrl: Invalid value`,
		},
		{
			name: "FailWebFolderWhenBaseUrlIsNotValid",
			obj: &deploymentsv1alpha1.Darkroom{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: deploymentsv1alpha1.DarkroomSpec{
					Source: deploymentsv1alpha1.Source{
						Type:          deploymentsv1alpha1.WebFolder,
						WebFolderMeta: deploymentsv1alpha1.WebFolderMeta{BaseURL: "invalid url"},
					},
					Domains: []string{"test.darkroom.com"},
				},
			},
			errString: `spec.source.baseUrl: Invalid value`,
		},
		{
			name: "FailS3WhenBucketIsMissing",
			obj: &deploymentsv1alpha1.Darkroom{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: deploymentsv1alpha1.DarkroomSpec{
					Source: deploymentsv1alpha1.Source{
						Type: deploymentsv1alpha1.S3,
					},
					Domains: []string{"test.darkroom.com"},
				},
			},
			errString: `spec.source.bucket: Required value: field required with Type S3`,
		},
		{
			name: "FailS3WhenBucketHasNameWithLengthLessThan3",
			obj: &deploymentsv1alpha1.Darkroom{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: deploymentsv1alpha1.DarkroomSpec{
					Source: deploymentsv1alpha1.Source{
						Type: deploymentsv1alpha1.S3,
						Bucket: &deploymentsv1alpha1.Bucket{
							Name: "ab",
						},
					},
					Domains: []string{"test.darkroom.com"},
				},
			},
			errString: `spec.source.bucket.name in body should be at least 3 chars long`,
		},
		{
			name: "FailS3WhenBucketHasNoAccessKey",
			obj: &deploymentsv1alpha1.Darkroom{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: deploymentsv1alpha1.DarkroomSpec{
					Source: deploymentsv1alpha1.Source{
						Type: deploymentsv1alpha1.S3,
						Bucket: &deploymentsv1alpha1.Bucket{
							Name: "test-bucket",
						},
					},
					Domains: []string{"test.darkroom.com"},
				},
			},
			errString: `spec.source.bucket.accessKey: Required value: field required with Type S3`,
		},
		{
			name: "FailS3WhenBucketHasNoSecretKey",
			obj: &deploymentsv1alpha1.Darkroom{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: deploymentsv1alpha1.DarkroomSpec{
					Source: deploymentsv1alpha1.Source{
						Type: deploymentsv1alpha1.S3,
						Bucket: &deploymentsv1alpha1.Bucket{
							Name:      "test-bucket",
							AccessKey: "some-key",
						},
					},
					Domains: []string{"test.darkroom.com"},
				},
			},
			errString: `spec.source.bucket.secretKey: Required value: field required with Type S3`,
		},
		{
			name: "FailGoogleCloudStorageWhenBucketIsMissing",
			obj: &deploymentsv1alpha1.Darkroom{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: deploymentsv1alpha1.DarkroomSpec{
					Source: deploymentsv1alpha1.Source{
						Type: deploymentsv1alpha1.GoogleCloudStorage,
					},
					Domains: []string{"test.darkroom.com"},
				},
			},
			errString: `spec.source.bucket: Required value: field required with Type GoogleCloudStorage`,
		},
		{
			name: "FailGoogleCloudStorageWhenBucketHasNoCredentialsJson",
			obj: &deploymentsv1alpha1.Darkroom{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: deploymentsv1alpha1.DarkroomSpec{
					Source: deploymentsv1alpha1.Source{
						Type: deploymentsv1alpha1.GoogleCloudStorage,
						Bucket: &deploymentsv1alpha1.Bucket{
							Name: "test-bucket",
						},
					},
					Domains: []string{"test.darkroom.com"},
				},
			},
			errString: `spec.source.bucket.credentialsJson: Required value: field required with Type GoogleCloudStorage`,
		},
		{
			name: "FailGoogleCloudStorageWithInvalidCredentialsJson",
			obj: &deploymentsv1alpha1.Darkroom{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				Spec: deploymentsv1alpha1.DarkroomSpec{
					Source: deploymentsv1alpha1.Source{
						Type: deploymentsv1alpha1.GoogleCloudStorage,
						Bucket: &deploymentsv1alpha1.Bucket{
							Name:            "test-bucket",
							CredentialsJson: "{some-bad-json-data}",
						},
					},
					Domains: []string{"test.darkroom.com"},
				},
			},
			errString: `spec.source.bucket.credentialsJson: Invalid value: "{some-bad-json-data}"`,
		},
	}

	for _, tc := range testcases {
		s.SetupTest()
		s.Run(tc.name, func() {
			err := s.client.Create(context.Background(), tc.obj, client.DryRunAll)
			s.Error(err)
			s.True(strings.Contains(err.Error(), tc.errString))
		})
	}
}
