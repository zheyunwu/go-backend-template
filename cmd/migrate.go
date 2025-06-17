package main

import (
	"log/slog"
	"strings"

	"github.com/go-backend-template/internal/di"
	"github.com/go-backend-template/internal/models"
)

// RunMigration is responsible for executing database migrations.
func RunMigration(env string) {
	slog.Info("Starting database migration...")

	// Initialize DI Container.
	diContainer := di.NewContainer(env)

	// Get database driver type.
	dbDriver := strings.ToLower(diContainer.Config.Database.Driver)
	slog.Info("Detected database driver", "driver", dbDriver)

	// Execute different migration strategies based on database type.
	var err error
	switch dbDriver {
	case "mysql":
		// MySQL specific table options.
		err = diContainer.DB.Set("gorm:table_options", "CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci").AutoMigrate(
			&models.User{},
			&models.UserProvider{},
			&models.Category{},
			&models.Product{},
			&models.ProductCategory{},
			&models.ProductImage{},
			&models.UserProductLike{},
			&models.UserProductFavorite{},
		)
	case "postgres", "postgresql":
		// PostgreSQL does not require specifying character sets, execute migration directly.
		err = diContainer.DB.AutoMigrate(
			&models.User{},
			&models.UserProvider{},
			&models.Category{},
			&models.Product{},
			&models.ProductCategory{},
			&models.ProductImage{},
			&models.UserProductLike{},
			&models.UserProductFavorite{},
		)
	default:
		slog.Error("Unsupported database driver for migration", "driver", dbDriver)
		return
	}

	if err != nil {
		slog.Error("Migration failed", "error", err)
		return
	}

	slog.Info("Database migration completed successfully!")
}
