package mocks

import (
	"context"

	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/internal/repositories"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/stretchr/testify/mock"
)

// MockProductRepository is a mock type for the ProductRepository type
type MockProductRepository struct {
	mock.Mock
}

// ListProducts provides a mock function with given fields: ctx, params
func (_m *MockProductRepository) ListProducts(ctx context.Context, params *query_params.QueryParams) ([]models.Product, int, error) {
	ret := _m.Called(ctx, params)

	var r0 []models.Product
	if rf, ok := ret.Get(0).(func(context.Context, *query_params.QueryParams) []models.Product); ok {
		r0 = rf(ctx, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Product)
		}
	}

	var r1 int
	if rf, ok := ret.Get(1).(func(context.Context, *query_params.QueryParams) int); ok {
		r1 = rf(ctx, params)
	} else {
		r1 = ret.Get(1).(int)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, *query_params.QueryParams) error); ok {
		r2 = rf(ctx, params)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetProduct provides a mock function with given fields: ctx, id
func (_m *MockProductRepository) GetProduct(ctx context.Context, id uint) (*models.Product, error) {
	ret := _m.Called(ctx, id)

	var r0 *models.Product
	if rf, ok := ret.Get(0).(func(context.Context, uint) *models.Product); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Product)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uint) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateProduct provides a mock function with given fields: ctx, product
func (_m *MockProductRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	ret := _m.Called(ctx, product)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.Product) error); ok {
		r0 = rf(ctx, product)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateProduct provides a mock function with given fields: ctx, id, updates
func (_m *MockProductRepository) UpdateProduct(ctx context.Context, id uint, updates map[string]interface{}) error {
	ret := _m.Called(ctx, id, updates)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uint, map[string]interface{}) error); ok {
		r0 = rf(ctx, id, updates)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteProduct provides a mock function with given fields: ctx, id
func (_m *MockProductRepository) DeleteProduct(ctx context.Context, id uint) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uint) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateProductWithRelations provides a mock function with given fields: ctx, product, images, categoryIDs
func (_m *MockProductRepository) CreateProductWithRelations(ctx context.Context, product *models.Product, images []models.ProductImage, categoryIDs []uint) error {
	ret := _m.Called(ctx, product, images, categoryIDs)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.Product, []models.ProductImage, []uint) error); ok {
		r0 = rf(ctx, product, images, categoryIDs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateProductWithRelations provides a mock function with given fields: ctx, id, updates, images, categoryIDs
func (_m *MockProductRepository) UpdateProductWithRelations(ctx context.Context, id uint, updates map[string]interface{}, images []models.ProductImage, categoryIDs []uint) error {
	ret := _m.Called(ctx, id, updates, images, categoryIDs)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uint, map[string]interface{}, []models.ProductImage, []uint) error); ok {
		r0 = rf(ctx, id, updates, images, categoryIDs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockProductRepository creates a new instance of MockProductRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockProductRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockProductRepository {
	mock := &MockProductRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}

var _ repositories.ProductRepository = (*MockProductRepository)(nil)
