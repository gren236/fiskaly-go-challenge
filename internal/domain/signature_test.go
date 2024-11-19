package domain

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockDeviceServer struct {
	mock.Mock
}

func (m *MockDeviceServer) GetDevice(ctx context.Context, deviceID uuid.UUID) (Device, error) {
	args := m.Called(ctx, deviceID)
	return args.Get(0).(Device), args.Error(1)
}

func (m *MockDeviceServer) IncrementSignatureCounter(ctx context.Context, deviceID uuid.UUID) error {
	args := m.Called(ctx, deviceID)
	return args.Error(0)
}

type MockSigner struct {
	mock.Mock
}

func (m *MockSigner) Sign(dataToBeSigned []byte) ([]byte, error) {
	args := m.Called(dataToBeSigned)
	signedData, ok := args.Get(0).([]byte)
	if !ok {
		return nil, args.Error(1)
	}
	return signedData, args.Error(1)
}

type MockSignerCreator struct {
	mock.Mock
}

func (m *MockSignerCreator) CreateSigner(kp KeyPair) (Signer, error) {
	args := m.Called(kp)
	signer, ok := args.Get(0).(Signer)
	if !ok {
		return nil, args.Error(1)
	}
	return signer, args.Error(1)
}

type MockSignaturePersister struct {
	mock.Mock
}

func (m *MockSignaturePersister) RunTransaction(ctx context.Context, deviceID uuid.UUID, fn func(ctx context.Context) error) error {
	args := m.Called(ctx, deviceID, fn)

	err := fn(ctx)
	if err != nil {
		return err
	}

	return args.Error(0)
}

func (m *MockSignaturePersister) SaveSignature(ctx context.Context, deviceID uuid.UUID, data SignedData) error {
	args := m.Called(ctx, deviceID, data)
	return args.Error(0)
}

func (m *MockSignaturePersister) GetLastSignature(ctx context.Context, deviceID uuid.UUID) (SignedData, error) {
	args := m.Called(ctx, deviceID)
	return args.Get(0).(SignedData), args.Error(1)
}

func (m *MockSignaturePersister) GetSignatures(ctx context.Context, deviceID uuid.UUID) ([]SignedData, error) {
	args := m.Called(ctx, deviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]SignedData), args.Error(1)
}

func TestSignatureService_SignTransaction_Success(t *testing.T) {
	logger := zap.NewNop().Sugar()
	deviceSvc := new(MockDeviceServer)
	signerCreator := new(MockSignerCreator)
	persister := new(MockSignaturePersister)

	deviceID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	device := Device{
		ID:               deviceID,
		SignatureCounter: 0,
		KeyPair:          &MockKeyPair{},
	}

	deviceSvc.On("GetDevice", mock.Anything, deviceID).Return(device, nil)
	signer := new(MockSigner)
	signerCreator.On("CreateSigner", device.KeyPair).Return(signer, nil)
	signer.On("Sign", mock.Anything).Return([]byte("signed_data"), nil)
	persister.On("RunTransaction", mock.Anything, deviceID, mock.Anything).Return(nil)
	persister.On("SaveSignature", mock.Anything, deviceID, mock.Anything).Return(nil)
	deviceSvc.On("IncrementSignatureCounter", mock.Anything, deviceID).Return(nil)

	ss := NewSignatureService(logger, deviceSvc, signerCreator, persister)
	signedData, err := ss.SignTransaction(context.Background(), deviceID, "data")

	assert.NoError(t, err)
	assert.Equal(t, "c2lnbmVkX2RhdGE=", signedData.Signature)
	assert.Equal(t, "0_data_MDAwMDAwMDAtMDAwMC0wMDAwLTAwMDAtMDAwMDAwMDAwMDAw", signedData.OriginalData)
}

func TestSignatureService_SignTransaction_GetDeviceError(t *testing.T) {
	logger := zap.NewNop().Sugar()
	deviceSvc := new(MockDeviceServer)
	signerCreator := new(MockSignerCreator)
	persister := new(MockSignaturePersister)

	deviceID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	deviceSvc.On("GetDevice", mock.Anything, deviceID).Return(Device{}, assert.AnError)
	signerCreator.On("CreateSigner", mock.Anything).Return(nil, nil)
	persister.On("RunTransaction", mock.Anything, deviceID, mock.Anything).Return(nil)

	ss := NewSignatureService(logger, deviceSvc, signerCreator, persister)
	_, err := ss.SignTransaction(context.Background(), deviceID, "data")

	assert.Error(t, err)
}

func TestSignatureService_SignTransaction_CreateSignerError(t *testing.T) {
	logger := zap.NewNop().Sugar()
	deviceSvc := new(MockDeviceServer)
	signerCreator := new(MockSignerCreator)
	persister := new(MockSignaturePersister)

	deviceID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	device := Device{
		ID:               deviceID,
		SignatureCounter: 0,
		KeyPair:          &MockKeyPair{},
	}

	deviceSvc.On("GetDevice", mock.Anything, deviceID).Return(device, nil)
	signerCreator.On("CreateSigner", device.KeyPair).Return(nil, assert.AnError)
	persister.On("RunTransaction", mock.Anything, deviceID, mock.Anything).Return(nil)

	ss := NewSignatureService(logger, deviceSvc, signerCreator, persister)
	_, err := ss.SignTransaction(context.Background(), deviceID, "data")

	assert.Error(t, err)
}

func TestSignatureService_SignTransaction_SignError(t *testing.T) {
	logger := zap.NewNop().Sugar()
	deviceSvc := new(MockDeviceServer)
	signerCreator := new(MockSignerCreator)
	persister := new(MockSignaturePersister)

	deviceID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	device := Device{
		ID:               deviceID,
		SignatureCounter: 0,
		KeyPair:          &MockKeyPair{},
	}

	deviceSvc.On("GetDevice", mock.Anything, deviceID).Return(device, nil)
	signer := new(MockSigner)
	signerCreator.On("CreateSigner", device.KeyPair).Return(signer, nil)
	signer.On("Sign", mock.Anything).Return(nil, assert.AnError)
	persister.On("RunTransaction", mock.Anything, deviceID, mock.Anything).Return(nil)

	ss := NewSignatureService(logger, deviceSvc, signerCreator, persister)
	_, err := ss.SignTransaction(context.Background(), deviceID, "data")

	assert.Error(t, err)
}

func TestSignatureService_SignTransaction_RunTransactionError(t *testing.T) {
	logger := zap.NewNop().Sugar()
	deviceSvc := new(MockDeviceServer)
	signerCreator := new(MockSignerCreator)
	persister := new(MockSignaturePersister)

	deviceID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	device := Device{
		ID:               deviceID,
		SignatureCounter: 0,
		KeyPair:          &MockKeyPair{},
	}

	deviceSvc.On("GetDevice", mock.Anything, deviceID).Return(device, nil)
	signer := new(MockSigner)
	signerCreator.On("CreateSigner", device.KeyPair).Return(signer, nil)
	signer.On("Sign", mock.Anything).Return([]byte("signed_data"), nil)
	persister.On("SaveSignature", mock.Anything, deviceID, mock.Anything).Return(nil)
	persister.On("RunTransaction", mock.Anything, deviceID, mock.Anything).Return(assert.AnError)
	deviceSvc.On("IncrementSignatureCounter", mock.Anything, deviceID).Return(nil)

	ss := NewSignatureService(logger, deviceSvc, signerCreator, persister)
	_, err := ss.SignTransaction(context.Background(), deviceID, "data")

	assert.Error(t, err)
}

func TestSignatureService_SignTransaction_SaveSignatureError(t *testing.T) {
	logger := zap.NewNop().Sugar()
	deviceSvc := new(MockDeviceServer)
	signerCreator := new(MockSignerCreator)
	persister := new(MockSignaturePersister)

	deviceID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	device := Device{
		ID:               deviceID,
		SignatureCounter: 0,
		KeyPair:          &MockKeyPair{},
	}

	deviceSvc.On("GetDevice", mock.Anything, deviceID).Return(device, nil)
	signer := new(MockSigner)
	signerCreator.On("CreateSigner", device.KeyPair).Return(signer, nil)
	signer.On("Sign", mock.Anything).Return([]byte("signed_data"), nil)
	persister.On("RunTransaction", mock.Anything, deviceID, mock.Anything).Return(nil)
	persister.On("SaveSignature", mock.Anything, deviceID, mock.Anything).Return(assert.AnError)
	deviceSvc.On("IncrementSignatureCounter", mock.Anything, deviceID).Return(nil)

	ss := NewSignatureService(logger, deviceSvc, signerCreator, persister)
	_, err := ss.SignTransaction(context.Background(), deviceID, "data")

	assert.Error(t, err)
}

func TestSignatureService_SignTransaction_IncrementSignatureCounterError(t *testing.T) {
	logger := zap.NewNop().Sugar()
	deviceSvc := new(MockDeviceServer)
	signerCreator := new(MockSignerCreator)
	persister := new(MockSignaturePersister)

	deviceID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	device := Device{
		ID:               deviceID,
		SignatureCounter: 0,
		KeyPair:          &MockKeyPair{},
	}

	deviceSvc.On("GetDevice", mock.Anything, deviceID).Return(device, nil)
	signer := new(MockSigner)
	signerCreator.On("CreateSigner", device.KeyPair).Return(signer, nil)
	signer.On("Sign", mock.Anything).Return([]byte("signed_data"), nil)
	persister.On("RunTransaction", mock.Anything, deviceID, mock.Anything).Return(nil)
	persister.On("SaveSignature", mock.Anything, deviceID, mock.Anything).Return(nil)
	deviceSvc.On("IncrementSignatureCounter", mock.Anything, deviceID).Return(assert.AnError)

	ss := NewSignatureService(logger, deviceSvc, signerCreator, persister)
	_, err := ss.SignTransaction(context.Background(), deviceID, "data")

	assert.Error(t, err)
}

func TestSignatureService_SignTransaction_GetLastSignatureError(t *testing.T) {
	logger := zap.NewNop().Sugar()
	deviceSvc := new(MockDeviceServer)
	signerCreator := new(MockSignerCreator)
	persister := new(MockSignaturePersister)

	deviceID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	device := Device{
		ID:               deviceID,
		SignatureCounter: 1,
		KeyPair:          &MockKeyPair{},
	}

	deviceSvc.On("GetDevice", mock.Anything, deviceID).Return(device, nil)
	signer := new(MockSigner)
	signerCreator.On("CreateSigner", device.KeyPair).Return(signer, nil)
	signer.On("Sign", mock.Anything).Return([]byte("signed_data"), nil)
	persister.On("RunTransaction", mock.Anything, deviceID, mock.Anything).Return(nil)
	persister.On("GetLastSignature", mock.Anything, deviceID).Return(SignedData{}, assert.AnError)

	ss := NewSignatureService(logger, deviceSvc, signerCreator, persister)
	_, err := ss.SignTransaction(context.Background(), deviceID, "data")

	assert.Error(t, err)
}

func TestSignatureService_SignTransaction_GetLastSignatureBaseCase(t *testing.T) {
	logger := zap.NewNop().Sugar()
	deviceSvc := new(MockDeviceServer)
	signerCreator := new(MockSignerCreator)
	persister := new(MockSignaturePersister)

	deviceID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	device := Device{
		ID:               deviceID,
		SignatureCounter: 1,
		KeyPair:          &MockKeyPair{},
	}

	deviceSvc.On("GetDevice", mock.Anything, deviceID).Return(device, nil)
	signer := new(MockSigner)
	signerCreator.On("CreateSigner", device.KeyPair).Return(signer, nil)
	signer.On("Sign", mock.Anything).Return([]byte("signed_data"), nil)
	persister.On("RunTransaction", mock.Anything, deviceID, mock.Anything).Return(nil)
	persister.On("GetLastSignature", mock.Anything, deviceID).Return(SignedData{}, nil)
	persister.On("SaveSignature", mock.Anything, deviceID, mock.Anything).Return(nil)
	deviceSvc.On("IncrementSignatureCounter", mock.Anything, deviceID).Return(nil)

	ss := NewSignatureService(logger, deviceSvc, signerCreator, persister)
	_, err := ss.SignTransaction(context.Background(), deviceID, "data")

	assert.NoError(t, err)
}

func TestSignatureService_GetSignatures(t *testing.T) {
	logger := zap.NewNop().Sugar()
	deviceSvc := new(MockDeviceServer)
	signerCreator := new(MockSignerCreator)
	persister := new(MockSignaturePersister)

	deviceID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	device := Device{
		ID:               deviceID,
		SignatureCounter: 1,
		KeyPair:          &MockKeyPair{},
	}

	deviceSvc.On("GetDevice", mock.Anything, deviceID).Return(device, nil)
	persister.On("GetSignatures", mock.Anything, deviceID).Return([]SignedData{{Signature: "signature"}}, nil)

	ss := NewSignatureService(logger, deviceSvc, signerCreator, persister)
	signatures, err := ss.GetSignatures(context.Background(), deviceID)

	assert.NoError(t, err)
	assert.Equal(t, []SignedData{{Signature: "signature"}}, signatures)
}

func TestSignatureService_GetSignatures_GetSignaturesError(t *testing.T) {
	logger := zap.NewNop().Sugar()
	deviceSvc := new(MockDeviceServer)
	signerCreator := new(MockSignerCreator)
	persister := new(MockSignaturePersister)

	deviceID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	device := Device{
		ID:               deviceID,
		SignatureCounter: 1,
		KeyPair:          &MockKeyPair{},
	}

	deviceSvc.On("GetDevice", mock.Anything, deviceID).Return(device, nil)
	persister.On("GetSignatures", mock.Anything, deviceID).Return(nil, assert.AnError)

	ss := NewSignatureService(logger, deviceSvc, signerCreator, persister)
	_, err := ss.GetSignatures(context.Background(), deviceID)

	assert.Error(t, err)
}
