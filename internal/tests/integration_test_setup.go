package tests

import (
	"net/http/httptest"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/di"
	"github.com/go-backend-template/internal/routes"
)

// SetupTestRouter initializes a new Gin engine with all routes and middlewares for testing.
// It uses the "test" environment for configuration.
func SetupTestRouter() (*gin.Engine, *di.Container) {
	// Set environment to "test"
	// Note: This might affect global state if not handled carefully.
	// Consider using per-test configurations or a dedicated test setup if this becomes an issue.
	os.Setenv("APP_ENV", "test")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	container := di.NewContainer("test") // Assumes "test" env uses in-memory or specific test DB config

	// Initialize routes
	routes.InitRoutes(router, container)

	return router, container
}

// PerformRequest is a helper function to execute an HTTP request against the test router.
func PerformRequest(r *gin.Engine, method, path string, body ...interface{}) *httptest.ResponseRecorder {
	// TODO: Handle request body if provided
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ClearTestData is a placeholder for clearing data from the test database.
// Actual implementation will depend on the database driver and ORM.
func ClearTestData(container *di.Container) {
	// Example for GORM:
	// container.DB.Exec("DELETE FROM products")
	// container.DB.Exec("DELETE FROM categories")
	// ... and so on for all relevant tables
	// For now, this is a no-op.
}

// SeedTestData is a placeholder for seeding data into the test database.
func SeedTestData(container *di.Container) {
	// Example for GORM:
	// category1 := models.Category{Name: "Electronics"}
	// container.DB.Create(&category1)
	// product1 := models.Product{Name: "Laptop", Price: 1200.00, CategoryID: category1.ID}
	// container.DB.Create(&product1)
	// For now, this is a no-op.
}
