package domain

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Algorithm string

const (
	AlgorithmECC Algorithm = "ECC"
	AlgorithmRSA Algorithm = "RSA"
)

func (a Algorithm) String() string {
	return string(a)
}

type KeyPair interface {
	IsKeyPair()
}

type Device struct {
	ID               uuid.UUID
	SignatureCounter uint64
	KeyPair          KeyPair
	Algorithm        Algorithm
	Label            *string
}

type DevicePersister interface {
	CreateDevice(ctx context.Context, device Device) error
	IncrementSignatureCounter(ctx context.Context, id uuid.UUID) error
	GetDevices(ctx context.Context) ([]Device, error)
	GetDevice(ctx context.Context, id uuid.UUID) (Device, error)
}

type KeyPairGenerator interface {
	GenerateKeyPair(algorithm Algorithm) (KeyPair, error)
}

type DeviceService struct {
	logger    *zap.SugaredLogger
	persister DevicePersister
	generator KeyPairGenerator
}

func NewDeviceService(logger *zap.SugaredLogger, persister DevicePersister, generator KeyPairGenerator) *DeviceService {
	return &DeviceService{
		logger:    logger,
		persister: persister,
		generator: generator,
	}
}

func (s *DeviceService) CreateDevice(ctx context.Context, label *string, algorithm Algorithm) (Device, error) {
	keyPair, err := s.generator.GenerateKeyPair(algorithm)
	if err != nil {
		return Device{}, err
	}

	device := Device{
		ID:               uuid.New(),
		SignatureCounter: 0,
		KeyPair:          keyPair,
		Algorithm:        algorithm,
		Label:            label,
	}

	err = s.persister.CreateDevice(ctx, device)
	if err != nil {
		return Device{}, err
	}

	s.logger.Infow("device created", "id", device.ID, "algorithm", device.Algorithm, "label", device.Label)

	return device, nil
}

func (s *DeviceService) IncrementSignatureCounter(ctx context.Context, id uuid.UUID) error {
	return s.persister.IncrementSignatureCounter(ctx, id)
}

func (s *DeviceService) GetDevices(ctx context.Context) ([]Device, error) {
	return s.persister.GetDevices(ctx)
}

func (s *DeviceService) GetDevice(ctx context.Context, id uuid.UUID) (Device, error) {
	return s.persister.GetDevice(ctx, id)
}
