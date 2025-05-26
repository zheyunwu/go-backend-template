package infra

import (
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/go-backend-template/config"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(config *config.Config) *gorm.DB {
	var DB *gorm.DB
	var err error
	dbConfig := config.Database

	// 根据数据库驱动类型选择不同的连接方式
	switch strings.ToLower(dbConfig.Driver) {
	case "mysql":
		DB, err = connectMySQL(dbConfig)
	case "postgres", "postgresql":
		DB, err = connectPostgres(dbConfig)
	default:
		log.Fatalf("Unsupported database driver: %s", dbConfig.Driver)
	}

	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// 设置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}

	// 设置连接池配置
	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)

	slog.Info("Database connected successfully", "driver", dbConfig.Driver)
	return DB
}

// 连接MySQL数据库
func connectMySQL(dbConfig config.DatabaseConfig) (*gorm.DB, error) {
	// 构建DSN (Data Source Name)
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&collation=utf8mb4_unicode_ci&parseTime=True&loc=UTC",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
		dbConfig.Charset,
	)

	slog.Info("Connecting to MySQL database", "host", dbConfig.Host, "port", dbConfig.Port, "database", dbConfig.Name)

	return gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
}

// 连接PostgreSQL数据库
func connectPostgres(dbConfig config.DatabaseConfig) (*gorm.DB, error) {
	// 构建PostgreSQL连接字符串
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		dbConfig.Host,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Name,
		dbConfig.Port,
	)

	slog.Info("Connecting to PostgreSQL database", "host", dbConfig.Host, "port", dbConfig.Port, "database", dbConfig.Name)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
}
