package main

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/di"
	"github.com/go-backend-template/internal/routes"
)

func StartServer(env string) {
	// Set Gin mode.
	if env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize DI Container.
	diContainer := di.NewContainer(env)

	// Create Gin instance.
	r := gin.New()

	// Initialize routes.
	routes.InitRoutes(r, diContainer)

	// Use configured port.
	port := strconv.Itoa(diContainer.Config.Server.Port)
	slog.Info("Server starting", "port", port, "env", env)
	slog.Info("Server running at http://localhost:" + port)

	if err := r.Run(":" + port); err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}
