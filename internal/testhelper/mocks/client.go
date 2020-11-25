package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	return m.Called(ctx, key, obj).Error(0)
}

func (m *MockClient) List(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
	return m.Called(ctx, list, opts).Error(0)
}

func (m *MockClient) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	return m.Called(ctx, obj, opts).Error(0)
}

func (m *MockClient) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	return m.Called(ctx, obj, opts).Error(0)
}

func (m *MockClient) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	return m.Called(ctx, obj, opts).Error(0)
}

func (m *MockClient) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	return m.Called(ctx, obj, patch, opts).Error(0)
}

func (m *MockClient) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...client.DeleteAllOfOption) error {
	return m.Called(ctx, obj, opts).Error(0)
}

func (m *MockClient) Status() client.StatusWriter {
	return m.Called().Get(0).(client.StatusWriter)
}
