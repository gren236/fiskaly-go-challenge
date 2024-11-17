package app

import (
	"context"
	"fmt"
	"github.com/fiskaly/coding-challenges/signing-service-challenge/pkg/config"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"os"
	"os/signal"
)

func Run(ctx context.Context, envGetter func(string) string, logger *zap.SugaredLogger) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// Introduce validator
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Parse config
	conf := NewConfig()
	err := ParseConfig(&conf, envGetter, validate)
	if err != nil {
		return err
	}

	return nil
}

func ParseConfig(conf any, getenv func(string) string, validate *validator.Validate) error {
	err := config.NewEnv(getenv).Set(conf)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	err = validate.Struct(conf)
	if err != nil {
		return fmt.Errorf("failed to validate config: %w", err)
	}

	return nil
}
