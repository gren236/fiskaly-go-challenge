package main

import (
	"context"
	"github.com/fiskaly/coding-challenges/signing-service-challenge/internal/app"
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
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			panic(err)
		}
	}(logger) // flushes buffer, if any

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
