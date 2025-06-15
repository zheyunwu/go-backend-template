package mocks

import (
	"context"

	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/internal/repositories"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/stretchr/testify/mock"
)

// MockCategoryRepository is a mock type for the CategoryRepository type
type MockCategoryRepository struct {
	mock.Mock
}

// ListCategories provides a mock function with given fields: ctx, params
func (_m *MockCategoryRepository) ListCategories(ctx context.Context, params *query_params.QueryParams) ([]models.Category, int, error) {
	ret := _m.Called(ctx, params)

	var r0 []models.Category
	if rf, ok := ret.Get(0).(func(context.Context, *query_params.QueryParams) []models.Category); ok {
		r0 = rf(ctx, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Category)
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

// GetCategory provides a mock function with given fields: ctx, id
func (_m *MockCategoryRepository) GetCategory(ctx context.Context, id uint) (*models.Category, error) {
	ret := _m.Called(ctx, id)

	var r0 *models.Category
	if rf, ok := ret.Get(0).(func(context.Context, uint) *models.Category); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Category)
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

// CreateCategory provides a mock function with given fields: ctx, category
func (_m *MockCategoryRepository) CreateCategory(ctx context.Context, category *models.Category) error {
	ret := _m.Called(ctx, category)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.Category) error); ok {
		r0 = rf(ctx, category)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateCategory provides a mock function with given fields: ctx, id, updates
func (_m *MockCategoryRepository) UpdateCategory(ctx context.Context, id uint, updates map[string]interface{}) error {
	ret := _m.Called(ctx, id, updates)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uint, map[string]interface{}) error); ok {
		r0 = rf(ctx, id, updates)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteCategory provides a mock function with given fields: ctx, id
func (_m *MockCategoryRepository) DeleteCategory(ctx context.Context, id uint) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uint) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetCategoryTree provides a mock function with given fields: ctx, depth, enabledOnly
func (_m *MockCategoryRepository) GetCategoryTree(ctx context.Context, depth int, enabledOnly bool) ([]models.Category, error) {
	ret := _m.Called(ctx, depth, enabledOnly)

	var r0 []models.Category
	if rf, ok := ret.Get(0).(func(context.Context, int, bool) []models.Category); ok {
		r0 = rf(ctx, depth, enabledOnly)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Category)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int, bool) error); ok {
		r1 = rf(ctx, depth, enabledOnly)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetChildCategories provides a mock function with given fields: ctx, parentID
func (_m *MockCategoryRepository) GetChildCategories(ctx context.Context, parentID uint) ([]models.Category, error) {
	ret := _m.Called(ctx, parentID)

	var r0 []models.Category
	if rf, ok := ret.Get(0).(func(context.Context, uint) []models.Category); ok {
		r0 = rf(ctx, parentID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Category)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uint) error); ok {
		r1 = rf(ctx, parentID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetCategoryProducts provides a mock function with given fields: ctx, categoryID, params
func (_m *MockCategoryRepository) GetCategoryProducts(ctx context.Context, categoryID uint, params *query_params.QueryParams) ([]models.Product, int, error) {
	ret := _m.Called(ctx, categoryID, params)

	var r0 []models.Product
	if rf, ok := ret.Get(0).(func(context.Context, uint, *query_params.QueryParams) []models.Product); ok {
		r0 = rf(ctx, categoryID, params)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Product)
		}
	}

	var r1 int
	if rf, ok := ret.Get(1).(func(context.Context, uint, *query_params.QueryParams) int); ok {
		r1 = rf(ctx, categoryID, params)
	} else {
		r1 = ret.Get(1).(int)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, uint, *query_params.QueryParams) error); ok {
		r2 = rf(ctx, categoryID, params)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// ExpandCategoryIDsWithChildren provides a mock function with given fields: ctx, categoryIDs
func (_m *MockCategoryRepository) ExpandCategoryIDsWithChildren(ctx context.Context, categoryIDs []uint) ([]uint, error) {
	ret := _m.Called(ctx, categoryIDs)

	var r0 []uint
	if rf, ok := ret.Get(0).(func(context.Context, []uint) []uint); ok {
		r0 = rf(ctx, categoryIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]uint)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []uint) error); ok {
		r1 = rf(ctx, categoryIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllChildCategoryIDs provides a mock function with given fields: ctx, parentID
func (_m *MockCategoryRepository) GetAllChildCategoryIDs(ctx context.Context, parentID uint) ([]uint, error) {
	ret := _m.Called(ctx, parentID)

	var r0 []uint
	if rf, ok := ret.Get(0).(func(context.Context, uint) []uint); ok {
		r0 = rf(ctx, parentID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]uint)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uint) error); ok {
		r1 = rf(ctx, parentID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewMockCategoryRepository creates a new instance of MockCategoryRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockCategoryRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCategoryRepository {
	mock := &MockCategoryRepository{}
	mock.Mock.Test(t)
	t.Cleanup(func() { mock.AssertExpectations(t) })
	return mock
}

var _ repositories.CategoryRepository = (*MockCategoryRepository)(nil)
