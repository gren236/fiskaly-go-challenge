package domain

import (
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SignedData struct {
	Signature    string // base64 encoded signature
	OriginalData string // original data used for signing
}

// Signer defines a contract for different types of signing implementations.
type Signer interface {
	Sign(dataToBeSigned []byte) ([]byte, error)
}

type SignerCreator interface {
	CreateSigner(kp KeyPair) (Signer, error)
}

type SignaturePersister interface {
	RunTransaction(ctx context.Context, deviceID uuid.UUID, fn func(ctx context.Context) error) error

	SaveSignature(ctx context.Context, deviceID uuid.UUID, data SignedData) error
	GetLastSignature(ctx context.Context, deviceID uuid.UUID) (SignedData, error)
	GetSignatures(ctx context.Context, deviceID uuid.UUID) ([]SignedData, error)
}

type SignatureService struct {
	logger        *zap.SugaredLogger
	deviceSvc     *DeviceService
	signerCreator SignerCreator
	persister     SignaturePersister
}

func NewSignatureService(logger *zap.SugaredLogger, deviceSvc *DeviceService, signerCreator SignerCreator, persister SignaturePersister) *SignatureService {
	return &SignatureService{
		logger:        logger,
		deviceSvc:     deviceSvc,
		signerCreator: signerCreator,
		persister:     persister,
	}
}

func (ss *SignatureService) SignTransaction(ctx context.Context, deviceID uuid.UUID, data string) (SignedData, error) {
	// TODO Get device
	// TODO Get last signature if device signature counter is not 0
	// TODO Sign data
	// TODO Save signature

	return SignedData{}, nil
}
