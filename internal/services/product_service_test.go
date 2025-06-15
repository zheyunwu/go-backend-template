package services_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-backend-template/internal/errors"
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/internal/repositories/mocks"
	"github.com/go-backend-template/internal/services"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/go-backend-template/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestProductService_GetProduct(t *testing.T) {
	mockProductRepo := mocks.NewMockProductRepository(t)
	mockCategoryRepo := mocks.NewMockCategoryRepository(t) // Needed for service instantiation
	productService := services.NewProductService(mockProductRepo, mockCategoryRepo)
	ctx := context.TODO()

	t.Run("Success", func(t *testing.T) {
		expectedProduct := &models.Product{ID: 1, Name: "Test Product"}
		mockProductRepo.On("GetProduct", ctx, uint(1)).Return(expectedProduct, nil).Once()

		product, err := productService.GetProduct(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, expectedProduct, product)
		// mockProductRepo.AssertExpectations(t) // Assertions are handled by NewMockXxxRepository helper
	})

	t.Run("Not Found", func(t *testing.T) {
		mockProductRepo.On("GetProduct", ctx, uint(2)).Return(nil, gorm.ErrRecordNotFound).Once()

		product, err := productService.GetProduct(ctx, 2)

		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Equal(t, errors.ErrProductNotFound, err)
	})

	t.Run("Other Error", func(t *testing.T) {
		genericError := fmt.Errorf("some database error")
		mockProductRepo.On("GetProduct", ctx, uint(3)).Return(nil, genericError).Once()

		product, err := productService.GetProduct(ctx, 3)

		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Contains(t, err.Error(), "failed to get product: some database error")
	})
}

func TestProductService_ListProducts(t *testing.T) {
	mockProductRepo := mocks.NewMockProductRepository(t)
	mockCategoryRepo := mocks.NewMockCategoryRepository(t)
	productService := services.NewProductService(mockProductRepo, mockCategoryRepo)
	ctx := context.TODO()
	params := &query_params.QueryParams{Page: 1, Limit: 10}

	t.Run("Success", func(t *testing.T) {
		expectedProducts := []models.Product{{ID: 1, Name: "Product 1"}, {ID: 2, Name: "Product 2"}}
		expectedTotal := 2
		mockProductRepo.On("ListProducts", ctx, params).Return(expectedProducts, expectedTotal, nil).Once()

		products, pagination, err := productService.ListProducts(ctx, params)

		assert.NoError(t, err)
		assert.Len(t, products, 2)
		assert.Equal(t, expectedProducts, products)
		assert.NotNil(t, pagination)
		assert.Equal(t, expectedTotal, pagination.TotalCount)
	})

	t.Run("Repository Error", func(t *testing.T) {
		repoError := fmt.Errorf("repo list error")
		mockProductRepo.On("ListProducts", ctx, params).Return(nil, 0, repoError).Once()

		products, pagination, err := productService.ListProducts(ctx, params)

		assert.Error(t, err)
		assert.Nil(t, products)
		assert.Nil(t, pagination)
		assert.Contains(t, err.Error(), "failed to list products: repo list error")
	})
}

func TestProductService_CreateProduct(t *testing.T) {
	mockProductRepo := mocks.NewMockProductRepository(t)
	mockCategoryRepo := mocks.NewMockCategoryRepository(t)
	productService := services.NewProductService(mockProductRepo, mockCategoryRepo)
	ctx := context.TODO()

	productToCreate := &models.Product{Name: "New Product"}
	images := []models.ProductImage{{ImageURL: "http://example.com/image.jpg"}}
	categoryIDs := []uint{1, 2}

	t.Run("Success", func(t *testing.T) {
		// Mock category checks
		mockCategoryRepo.On("GetCategory", ctx, uint(1)).Return(&models.Category{ID: 1}, nil).Once()
		mockCategoryRepo.On("GetCategory", ctx, uint(2)).Return(&models.Category{ID: 2}, nil).Once()

		// Mock product creation
		// We need to ensure the product passed to CreateProductWithRelations has its ID set by the mock
		// or that the service correctly returns the ID from the input product model after creation.
		// For this test, we'll assume CreateProductWithRelations sets the ID on the passed product.
		mockProductRepo.On("CreateProductWithRelations", ctx, mock.AnythingOfType("*models.Product"), images, categoryIDs).
			Run(func(args mock.Arguments) {
				argProduct := args.Get(1).(*models.Product)
				argProduct.ID = 123 // Simulate DB assigning an ID
			}).Return(nil).Once()


		createdID, err := productService.CreateProduct(ctx, productToCreate, images, categoryIDs)

		assert.NoError(t, err)
		assert.Equal(t, uint(123), createdID) // Check if the service returned the ID
	})

	t.Run("Validation - Name Empty", func(t *testing.T) {
		emptyNameProduct := &models.Product{Name: ""}
		createdID, err := productService.CreateProduct(ctx, emptyNameProduct, images, categoryIDs)

		assert.Error(t, err)
		assert.Equal(t, uint(0), createdID)
		assert.Equal(t, errors.ErrProductNameEmpty, err)
	})

	t.Run("Category Not Found", func(t *testing.T) {
		mockCategoryRepo.On("GetCategory", ctx, uint(1)).Return(&models.Category{ID: 1}, nil).Once()
		mockCategoryRepo.On("GetCategory", ctx, uint(3)).Return(nil, gorm.ErrRecordNotFound).Once() // Category 3 not found

		createdID, err := productService.CreateProduct(ctx, productToCreate, images, []uint{1, 3})

		assert.Error(t, err)
		assert.Equal(t, uint(0), createdID)
		assert.Equal(t, errors.ErrCategoryNotFound, err)
	})

	t.Run("Category Repo Other Error", func(t *testing.T) {
		repoError := fmt.Errorf("category repo error")
		mockCategoryRepo.On("GetCategory", ctx, uint(1)).Return(nil, repoError).Once()

		createdID, err := productService.CreateProduct(ctx, productToCreate, images, []uint{1})

		assert.Error(t, err)
		assert.Equal(t, uint(0), createdID)
		assert.Contains(t, err.Error(), "failed to check category 1: category repo error")
	})

	t.Run("Product Creation Repo Error", func(t *testing.T) {
		mockCategoryRepo.On("GetCategory", ctx, uint(1)).Return(&models.Category{ID: 1}, nil).Once()
		mockCategoryRepo.On("GetCategory", ctx, uint(2)).Return(&models.Category{ID: 2}, nil).Once()

		repoError := fmt.Errorf("product creation repo error")
		mockProductRepo.On("CreateProductWithRelations", ctx, productToCreate, images, categoryIDs).Return(repoError).Once()

		createdID, err := productService.CreateProduct(ctx, productToCreate, images, categoryIDs)

		assert.Error(t, err)
		assert.Equal(t, uint(0), createdID)
		assert.Contains(t, err.Error(), "failed to create product: product creation repo error")
	})
}

// Note: Tests for UpdateProduct and DeleteProduct would follow a similar pattern,
// mocking repository calls and asserting service behavior based on different return values (success, not found, other errors).
// For UpdateProduct, one would also test validation logic (e.g., product name empty if provided in updates).
// For brevity, these are not fully implemented here but the structure would be similar to CreateProduct.
