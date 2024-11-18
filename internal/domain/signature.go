package domain

import (
	"context"
	"encoding/base64"
	"fmt"
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
	var signedData *SignedData

	err := ss.persister.RunTransaction(ctx, deviceID, func(ctx context.Context) error {
		// Get device
		device, err := ss.deviceSvc.GetDevice(ctx, deviceID)
		if err != nil {
			return err
		}

		// Get last signature if device signature counter is not 0
		lastSignature := base64.StdEncoding.EncodeToString([]byte(device.ID.String())) // base case
		if device.SignatureCounter != 0 {
			lastSignatureData, err := ss.persister.GetLastSignature(ctx, deviceID)
			if err != nil {
				return fmt.Errorf("failed to retrieve last signature: %w", err)
			}

			lastSignature = lastSignatureData.Signature
		}

		// Sign data
		signer, err := ss.signerCreator.CreateSigner(device.KeyPair)
		if err != nil {
			return fmt.Errorf("failed to create signer: %w", err)
		}

		dataToBeSigned := fmt.Sprintf("%d_%s_%s", device.SignatureCounter, data, lastSignature)

		signature, err := signer.Sign([]byte(dataToBeSigned))
		if err != nil {
			return fmt.Errorf("failed to sign data: %w", err)
		}

		// Save signature
		signedData = &SignedData{
			Signature:    base64.StdEncoding.EncodeToString(signature),
			OriginalData: dataToBeSigned,
		}

		err = ss.persister.SaveSignature(ctx, deviceID, *signedData)
		if err != nil {
			return fmt.Errorf("failed to save signature: %w", err)
		}

		// Update device signature counter
		err = ss.deviceSvc.IncrementSignatureCounter(ctx, deviceID)
		if err != nil {
			return fmt.Errorf("failed to increment signature counter: %w", err)
		}

		return nil
	})
	if err != nil {
		return SignedData{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return *signedData, nil
}

func (ss *SignatureService) GetSignatures(ctx context.Context, deviceID uuid.UUID) ([]SignedData, error) {
	signatures, err := ss.persister.GetSignatures(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve signatures: %w", err)
	}

	return signatures, nil
}
