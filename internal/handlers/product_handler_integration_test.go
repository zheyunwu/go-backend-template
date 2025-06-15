package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-backend-template/internal/dto"
	"github.com/go-backend-template/internal/tests" // Import the test helpers
	"github.com/go-backend-template/pkg/response"
	"github.com/stretchr/testify/assert"
)

func TestProductAPI_ListProducts_Integration(t *testing.T) {
	router, container := tests.SetupTestRouter()
	// Defer cleanup if necessary, e.g., closing DB connections or clearing data
	// defer tests.ClearTestData(container) // Implement this if using a persistent test DB

	// Seed initial data if needed for this specific test
	// For now, we expect an empty list or whatever is in the default DB
	// tests.SeedTestData(container) // Example: Seed some products

	t.Run("ListProducts - Basic", func(t *testing.T) {
		w := tests.PerformRequest(router, "GET", "/api/v1/products")

		assert.Equal(t, http.StatusOK, w.Code)

		var resp response.SuccessResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		assert.True(t, resp.Success)
		assert.Equal(t, "Successfully retrieved data.", resp.Message) // Default success message

		// Assert data type is a list of products (or an empty list)
		// We need to be careful with type assertion if data can be nil
		if resp.Data != nil {
			_, ok := resp.Data.([]interface{}) // Gin often decodes to []interface{}
			// If you have a more specific DTO slice, you might need a more complex assertion
			// or re-marshal and unmarshal into the specific slice type.
			// For this basic test, checking if it's a slice is a good start.
			// A more robust test would check the actual DTO structure.
			assert.True(t, ok, "Data should be a slice of products")

			// Example of checking actual DTOs if data is not empty:
			// if ok && len(dataSlice) > 0 {
			// 	var productsDTO []dto.UserProductDTO
			//  // This requires converting []interface{} to []dto.UserProductDTO
			//  // which can be done by ranging over dataSlice and type asserting each element
			//  // or by re-marshalling `resp.Data` and unmarshalling to `&productsDTO`
			// 	jsonData, _ := json.Marshal(resp.Data)
			// 	json.Unmarshal(jsonData, &productsDTO)
			// 	assert.NotEmpty(t, productsDTO)
			// }

		} else {
			// If Data is nil, it implies an empty list was correctly returned as nil (or an empty slice)
			// Depending on how your response package marshals empty slices (as `null` or `[]`)
			// this assertion might need adjustment.
			// For now, we accept nil as a valid representation of no data.
		}

		// Assert pagination structure
		assert.NotNil(t, resp.Pagination)
		assert.GreaterOrEqual(t, resp.Pagination.TotalCount, 0)
		assert.GreaterOrEqual(t, resp.Pagination.PageSize, 0) // Or specific default
		assert.GreaterOrEqual(t, resp.Pagination.CurrentPage, 0) // Or specific default
		assert.GreaterOrEqual(t, resp.Pagination.TotalPages, 0)
	})

	// TODO: Add more tests with query parameters (limit, page, search, filter)
	// These would require seeding data and then asserting the filtered/paginated results.
	// Example:
	// t.Run("ListProducts - With Pagination", func(t *testing.T) {
	// 	// 1. Seed more products than default page size
	// 	// 2. Perform request with ?page=2&limit=5
	// 	// 3. Assert correct subset of products and pagination details
	// })
}

// Note: A more complete integration test suite would involve:
// - Setting up a dedicated test database.
// - Implementing SeedTestData and ClearTestData to manage DB state between tests.
// - Testing various scenarios: empty list, single item, multiple items, pagination, filtering, searching.
// - For POST/PATCH/DELETE endpoints, verifying data changes in the database.
// - Testing authentication and authorization for protected endpoints.
// - Testing error responses (4xx, 5xx) for invalid inputs or server errors.
