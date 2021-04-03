package v1alpha1

import (
	"reflect"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestDarkroom_Default(t *testing.T) {
	type fields struct {
		Spec DarkroomSpec
	}
	tests := []struct {
		name string
		obj  Darkroom
		want fields
	}{
		{
			name: "DefaultVersion",
			obj:  Darkroom{},
			want: fields{Spec: DarkroomSpec{
				Version: "latest",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.obj.Default()
			if !reflect.DeepEqual(tt.want.Spec, tt.obj.Spec) {
				t.Errorf("Default() got = %v, want %v", tt.obj.Spec, tt.want.Spec)
			}
		})
	}
}

func TestDarkroom_ValidateCreate(t *testing.T) {
	type fields struct {
		Spec DarkroomSpec
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "WebFolderCreate",
			fields: fields{
				Spec: DarkroomSpec{
					Source: Source{
						Type:          WebFolder,
						WebFolderMeta: WebFolderMeta{BaseURL: "https://example.com"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "S3Create",
			fields: fields{
				Spec: DarkroomSpec{
					Source: Source{
						Type:   S3,
						Bucket: &Bucket{AccessKey: "access", SecretKey: "secret"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "GoogleCloudStorageCreate",
			fields: fields{
				Spec: DarkroomSpec{
					Source: Source{
						Type:   GoogleCloudStorage,
						Bucket: &Bucket{CredentialsJson: "{}"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "WebFolderHasNoBaseUrl",
			fields: fields{
				Spec: DarkroomSpec{
					Source: Source{
						Type:          WebFolder,
						WebFolderMeta: WebFolderMeta{BaseURL: "random"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "S3HasNoBucket",
			fields: fields{
				Spec: DarkroomSpec{
					Source: Source{
						Type: S3,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "S3HasNoAccessKey",
			fields: fields{
				Spec: DarkroomSpec{
					Source: Source{
						Type:   S3,
						Bucket: &Bucket{AccessKey: "access"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "S3HasNoSecretKey",
			fields: fields{
				Spec: DarkroomSpec{
					Source: Source{
						Type:   S3,
						Bucket: &Bucket{SecretKey: "secret"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "GoogleCloudStorageHasNoBucket",
			fields: fields{
				Spec: DarkroomSpec{
					Source: Source{
						Type: GoogleCloudStorage,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "GoogleCloudStorageHasNoCredentialsJson",
			fields: fields{
				Spec: DarkroomSpec{
					Source: Source{
						Type:   GoogleCloudStorage,
						Bucket: &Bucket{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "GoogleCloudStorageHasInvalidCredentialsJson",
			fields: fields{
				Spec: DarkroomSpec{
					Source: Source{
						Type:   GoogleCloudStorage,
						Bucket: &Bucket{CredentialsJson: "random"},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Darkroom{
				Spec: tt.fields.Spec,
			}
			if err := d.ValidateCreate(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDarkroom_ValidateUpdate(t *testing.T) {
	tm := v1.TypeMeta{
		Kind:       "Darkroom",
		APIVersion: GroupVersion.String(),
	}
	om := v1.ObjectMeta{
		Name:      "sample",
		Namespace: "default",
	}
	type fields struct {
		TypeMeta   v1.TypeMeta
		ObjectMeta v1.ObjectMeta
		Spec       DarkroomSpec
	}
	type args struct {
		old runtime.Object
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Success",
			fields: fields{
				TypeMeta:   tm,
				ObjectMeta: om,
				Spec:       DarkroomSpec{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Darkroom{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
			}
			if err := d.ValidateUpdate(tt.args.old); (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDarkroom_ValidateDelete(t *testing.T) {
	tm := v1.TypeMeta{
		Kind:       "Darkroom",
		APIVersion: GroupVersion.String(),
	}
	om := v1.ObjectMeta{
		Name:      "sample",
		Namespace: "default",
	}
	type fields struct {
		TypeMeta   v1.TypeMeta
		ObjectMeta v1.ObjectMeta
		Spec       DarkroomSpec
		Status     DarkroomStatus
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Success",
			fields: fields{
				TypeMeta:   tm,
				ObjectMeta: om,
				Spec:       DarkroomSpec{},
				Status:     DarkroomStatus{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Darkroom{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			if err := d.ValidateDelete(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
