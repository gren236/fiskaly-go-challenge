package main

import (
	"context"
	"github.com/gren236/fiskaly-go-challenge/internal/app"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"os"
)

func main() {
	// Introduce logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync() // nolint:errcheck

	sLogger := logger.Sugar()

	err = godotenv.Load()
	if err != nil {
		sLogger.Warn(".env file not found")
	}

	ctx := context.Background()

	if err = app.Run(ctx, os.Getenv, sLogger); err != nil {
		sLogger.Fatal(err)
	}
}
