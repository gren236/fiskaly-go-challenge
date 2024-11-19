package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gren236/fiskaly-go-challenge/internal/api"
	"github.com/gren236/fiskaly-go-challenge/internal/crypto"
	"github.com/gren236/fiskaly-go-challenge/internal/domain"
	"github.com/gren236/fiskaly-go-challenge/internal/persistence"
	"github.com/gren236/fiskaly-go-challenge/pkg/config"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
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

	logger.Infow("parsed config", "config", conf)

	// Set up crypto services
	keyGenerator := crypto.NewGenerator()
	signerCreator := crypto.NewSignerCreator()
	kpMarshaler := crypto.NewMarshaler()

	// Set up persistence
	inMemory := persistence.NewInMemory(kpMarshaler)

	// Set up services
	deviceService := domain.NewDeviceService(logger, inMemory, keyGenerator)
	signatureService := domain.NewSignatureService(logger, deviceService, signerCreator, inMemory)

	// Set up the server. I've extended the server setup so we could gracefully shutdown it with context cancellation.
	server := api.NewServer(logger, api.Config{Host: conf.ApiHost, Port: conf.ApiPort}, validate, deviceService, signatureService)

	logger.Info("built all dependencies")

	// Set up wait group for all goroutines
	var wg sync.WaitGroup

	s := server.GetHttpServer()

	go func() {
		logger.Infof("listening on %s", s.Addr)

		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(fmt.Errorf("error starting http server: %w", err))
		}
	}()

	wg.Add(1)

	// Set up graceful shutdown of the server
	go func() {
		defer wg.Done()

		<-ctx.Done()

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutdown()

		if err := s.Shutdown(shutdownCtx); err != nil {
			logger.Error(fmt.Errorf("error shutting down http server: %w", err))
		}

		logger.Info("api server shutdown successful")
	}()

	wg.Wait()

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
