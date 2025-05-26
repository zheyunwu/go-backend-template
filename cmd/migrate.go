package main

import (
	"log/slog"
	"strings"

	"github.com/go-backend-template/internal/di"
	"github.com/go-backend-template/internal/models"
)

// RunMigration 负责执行数据库迁移
func RunMigration(env string) {
	slog.Info("Starting database migration...")

	// 初始化DI Container
	diContainer := di.NewContainer(env)

	// 获取数据库驱动类型
	dbDriver := strings.ToLower(diContainer.Config.Database.Driver)
	slog.Info("Detected database driver", "driver", dbDriver)

	// 根据数据库类型执行不同的迁移策略
	var err error
	switch dbDriver {
	case "mysql":
		// MySQL特有的表选项
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
		// PostgreSQL不需要指定字符集，直接执行迁移
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
