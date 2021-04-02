package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MockClient struct {
	mock.Mock
	RuntimeScheme *runtime.Scheme
}

func (m *MockClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return m.Called(ctx, key, obj).Error(0)
}

func (m *MockClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return m.Called(ctx, list, opts).Error(0)
}

func (m *MockClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return m.Called(ctx, obj, opts).Error(0)
}

func (m *MockClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return m.Called(ctx, obj, opts).Error(0)
}

func (m *MockClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return m.Called(ctx, obj, opts).Error(0)
}

func (m *MockClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return m.Called(ctx, obj, patch, opts).Error(0)
}

func (m *MockClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return m.Called(ctx, obj, opts).Error(0)
}

func (m *MockClient) Status() client.StatusWriter {
	return m.Called().Get(0).(client.StatusWriter)
}

func (m *MockClient) Scheme() *runtime.Scheme {
	return m.RuntimeScheme
}

func (m *MockClient) RESTMapper() meta.RESTMapper {
	return &restMapper{}
}

type restMapper struct {
	mock.Mock
}

func (r *restMapper) KindFor(resource schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	panic("implement me")
}

func (r *restMapper) KindsFor(resource schema.GroupVersionResource) ([]schema.GroupVersionKind, error) {
	panic("implement me")
}

func (r *restMapper) ResourceFor(input schema.GroupVersionResource) (schema.GroupVersionResource, error) {
	panic("implement me")
}

func (r *restMapper) ResourcesFor(input schema.GroupVersionResource) ([]schema.GroupVersionResource, error) {
	panic("implement me")
}

func (r *restMapper) RESTMapping(gk schema.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	panic("implement me")
}

func (r *restMapper) RESTMappings(gk schema.GroupKind, versions ...string) ([]*meta.RESTMapping, error) {
	panic("implement me")
}

func (r *restMapper) ResourceSingularizer(resource string) (singular string, err error) {
	panic("implement me")
}
