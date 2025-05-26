package main

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/di"
	"github.com/go-backend-template/internal/routes"
)

func StartServer(env string) {
	// 设置 Gin 模式
	if env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 初始化DI Container
	diContainer := di.NewContainer(env)

	// 创建 Gin 实例
	r := gin.New()

	// 初始化路由
	routes.InitRoutes(r, diContainer)

	// 使用配置的端口
	port := strconv.Itoa(diContainer.Config.Server.Port)
	slog.Info("Server starting", "port", port, "env", env)
	slog.Info("Server running at http://localhost:" + port)

	if err := r.Run(":" + port); err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}
