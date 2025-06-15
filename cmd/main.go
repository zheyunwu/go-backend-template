package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/go-backend-template/pkg/logger"
)

func main() {
	// Initialize logger
	logger.Init(&logger.Config{
		Level:      slog.LevelDebug,
		JSONFormat: true,
		Output:     os.Stdout,
	})

	time.Local = time.UTC

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: main [server|migrate]")
		return
	}

	switch os.Args[1] {
	case "server":
		StartServer(env)
	case "migrate":
		RunMigration(env)
	default:
		slog.Error("Unknown command. Use 'server' or 'migrate'.")
	}
}
