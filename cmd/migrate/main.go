package main

import (
	"context"
	"fmt"
	"os"

	"deeler/storage"

	"github.com/maragudk/env"
	"github.com/maragudk/migrate"
	"go.uber.org/zap"
)

func main() {
	os.Exit(start())
}

// Utility function that connects to the database and apply migrations.
func start() int {
	_ = env.Load(".env")

	logEnv := env.GetStringOrDefault("LOG_ENV", "development") // Default to development
	log, err := createLogger(logEnv)

	if err != nil {
		fmt.Println("Failed to create logger:", err)
	}

	if len(os.Args) < 2 {
		log.Warn("Usage: migrate up|down|to")
		return 1
	}

	if os.Args[1] == "to" && len(os.Args) < 3 {
		log.Warn("Usage: migrate to <version>")
		return 1
	}

	// Initiate a new database connection.
	db := storage.NewDatabase(storage.NewDatabaseOptions{
		Host:     env.GetStringOrDefault("DB_HOST", "localhost"),
		Port:     env.GetIntOrDefault("DB_PORT", 5432),
		User:     env.GetStringOrDefault("DB_USER", ""),
		Password: env.GetStringOrDefault("DB_PASSWORD", ""),
		Name:     env.GetStringOrDefault("DB_NAME", ""),
	})

	if err := db.Connect(); err != nil {
		log.Error("Failed to connect to database", zap.Error(err))
		return 1
	}

	fsys := os.DirFS("storage/migrations")
	switch os.Args[1] {
	case "up":
		err = migrate.Up(context.Background(), db.DB.DB, fsys)
	case "down":
		err = migrate.Down(context.Background(), db.DB.DB, fsys)
	case "to":
		err = migrate.To(context.Background(), db.DB.DB, fsys, os.Args[2])
	default:
		log.Error("Unknown command", zap.String("name", os.Args[1]))
		return 1
	}
	if err != nil {
		log.Error("Error migrating", zap.Error(err))
		return 1
	}

	return 0
}

func createLogger(env string) (*zap.Logger, error) {
	switch env {
	case "production":
		return zap.NewProduction()
	case "development":
		return zap.NewDevelopment()
	default:
		return zap.NewNop(), nil
	}
}
