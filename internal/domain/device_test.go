package domain

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockDevicePersister struct {
	mock.Mock
}

func (m *MockDevicePersister) CreateDevice(ctx context.Context, device Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDevicePersister) IncrementSignatureCounter(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDevicePersister) GetDevices(ctx context.Context) ([]Device, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Device), args.Error(1)
}

func (m *MockDevicePersister) GetDevice(ctx context.Context, id uuid.UUID) (Device, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Device), args.Error(1)
}

type MockKeyPairGenerator struct {
	mock.Mock
}

func (m *MockKeyPairGenerator) GenerateKeyPair(algorithm Algorithm) (KeyPair, error) {
	args := m.Called(algorithm)
	if args.Get(0) == nil {
		return &MockKeyPair{}, args.Error(1)
	}
	return args.Get(0).(KeyPair), args.Error(1)
}

type MockKeyPair struct {
	mock.Mock
}

func (m *MockKeyPair) IsKeyPair() {}

func TestDeviceService_CreateDevice_Success(t *testing.T) {
	logger := zap.NewNop().Sugar()
	persister := new(MockDevicePersister)
	generator := new(MockKeyPairGenerator)
	service := NewDeviceService(logger, persister, generator)

	ctx := context.Background()
	label := "test-device"
	algorithm := AlgorithmRSA
	keyPair := new(MockKeyPair)

	generator.On("GenerateKeyPair", algorithm).Return(keyPair, nil)
	persister.On("CreateDevice", ctx, mock.AnythingOfType("Device")).Return(nil)

	device, err := service.CreateDevice(ctx, &label, algorithm)

	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, algorithm, device.Algorithm)
	assert.Equal(t, &label, device.Label)
	generator.AssertExpectations(t)
	persister.AssertExpectations(t)
}

func TestDeviceService_CreateDevice_GenerateKeyPairError(t *testing.T) {
	logger := zap.NewNop().Sugar()
	persister := new(MockDevicePersister)
	generator := new(MockKeyPairGenerator)
	service := NewDeviceService(logger, persister, generator)

	ctx := context.Background()
	label := "test-device"
	algorithm := AlgorithmRSA

	generator.On("GenerateKeyPair", algorithm).Return(nil, errors.New("key pair generation error"))

	device, err := service.CreateDevice(ctx, &label, algorithm)

	assert.Error(t, err)
	assert.EqualError(t, err, "key pair generation error")
	assert.Equal(t, Device{}, device)
	generator.AssertExpectations(t)
	persister.AssertExpectations(t)
}

func TestDeviceService_CreateDevice_PersisterError(t *testing.T) {
	logger := zap.NewNop().Sugar()
	persister := new(MockDevicePersister)
	generator := new(MockKeyPairGenerator)
	service := NewDeviceService(logger, persister, generator)

	ctx := context.Background()
	label := "test-device"
	algorithm := AlgorithmRSA
	keyPair := new(MockKeyPair)

	generator.On("GenerateKeyPair", algorithm).Return(keyPair, nil)
	persister.On("CreateDevice", ctx, mock.AnythingOfType("Device")).Return(errors.New("persister error"))

	device, err := service.CreateDevice(ctx, &label, algorithm)

	assert.Error(t, err)
	assert.EqualError(t, err, "persister error")
	assert.Equal(t, Device{}, device)
	generator.AssertExpectations(t)
	persister.AssertExpectations(t)
}

func TestDeviceService_IncrementSignatureCounter_Success(t *testing.T) {
	logger := zap.NewNop().Sugar()
	persister := new(MockDevicePersister)
	service := NewDeviceService(logger, persister, nil)

	ctx := context.Background()
	id := uuid.New()

	persister.On("IncrementSignatureCounter", ctx, id).Return(nil)

	err := service.IncrementSignatureCounter(ctx, id)

	assert.NoError(t, err)
	persister.AssertExpectations(t)
}

func TestDeviceService_IncrementSignatureCounter_PersisterError(t *testing.T) {
	logger := zap.NewNop().Sugar()
	persister := new(MockDevicePersister)
	service := NewDeviceService(logger, persister, nil)

	ctx := context.Background()
	id := uuid.New()

	persister.On("IncrementSignatureCounter", ctx, id).Return(errors.New("persister error"))

	err := service.IncrementSignatureCounter(ctx, id)

	assert.Error(t, err)
	assert.EqualError(t, err, "persister error")
	persister.AssertExpectations(t)
}

func TestDeviceService_GetDevices_Success(t *testing.T) {
	logger := zap.NewNop().Sugar()
	persister := new(MockDevicePersister)
	service := NewDeviceService(logger, persister, nil)

	ctx := context.Background()
	devices := []Device{
		{ID: uuid.New(), Algorithm: AlgorithmRSA},
		{ID: uuid.New(), Algorithm: AlgorithmECC},
	}

	persister.On("GetDevices", ctx).Return(devices, nil)

	result, err := service.GetDevices(ctx)

	assert.NoError(t, err)
	assert.Equal(t, devices, result)
	persister.AssertExpectations(t)
}

func TestDeviceService_GetDevices_PersisterError(t *testing.T) {
	logger := zap.NewNop().Sugar()
	persister := new(MockDevicePersister)
	service := NewDeviceService(logger, persister, nil)

	ctx := context.Background()

	persister.On("GetDevices", ctx).Return(nil, errors.New("persister error"))

	result, err := service.GetDevices(ctx)

	assert.Error(t, err)
	assert.EqualError(t, err, "persister error")
	assert.Nil(t, result)
	persister.AssertExpectations(t)
}

func TestDeviceService_GetDevice_Success(t *testing.T) {
	logger := zap.NewNop().Sugar()
	persister := new(MockDevicePersister)
	service := NewDeviceService(logger, persister, nil)

	ctx := context.Background()
	id := uuid.New()
	device := Device{ID: id, Algorithm: AlgorithmRSA}

	persister.On("GetDevice", ctx, id).Return(device, nil)

	result, err := service.GetDevice(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, device, result)
	persister.AssertExpectations(t)
}

func TestDeviceService_GetDevice_PersisterError(t *testing.T) {
	logger := zap.NewNop().Sugar()
	persister := new(MockDevicePersister)
	service := NewDeviceService(logger, persister, nil)

	ctx := context.Background()
	id := uuid.New()

	persister.On("GetDevice", ctx, id).Return(Device{}, errors.New("persister error"))

	result, err := service.GetDevice(ctx, id)

	assert.Error(t, err)
	assert.EqualError(t, err, "persister error")
	assert.Equal(t, Device{}, result)
	persister.AssertExpectations(t)
}
