package persistence

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/gren236/fiskaly-go-challenge/internal/domain"
	"sync"
	"time"
)

type KeyPairMarshaler interface {
	Marshal(pair domain.KeyPair) ([]byte, []byte, error)
	Unmarshal(algo domain.Algorithm, privateKeyBytes []byte) (domain.KeyPair, error)
}

type Signature struct {
	signature    string // base64 encoded signature
	originalData string // original data used for signing
	createdAt    time.Time
}

type Device struct {
	id               uuid.UUID
	signatureCounter uint64
	privateKey       []byte
	algorithm        string
	label            *string
	signatures       []Signature

	sync.Mutex // We need to lock the device when adding a new signature
}

// InMemory is an in-memory implementation of the persistence layer.
type InMemory struct {
	storage map[uuid.UUID]*Device

	kpMarshaler KeyPairMarshaler
}

// NewInMemory creates a new InMemory persistence layer. I pass context to every function to be able to cancel the
// operation if needed. This is a good practice, even if we do not use it in this implementation.
func NewInMemory(kpMarshaler KeyPairMarshaler) *InMemory {
	return &InMemory{
		storage:     make(map[uuid.UUID]*Device),
		kpMarshaler: kpMarshaler,
	}
}

// CreateDevice creates a new device in the persistence layer.
func (p *InMemory) CreateDevice(ctx context.Context, device domain.Device) error {
	_, priv, err := p.kpMarshaler.Marshal(device.KeyPair)
	if err != nil {
		return fmt.Errorf("could not marshal key pair: %w", err)
	}

	p.storage[device.ID] = &Device{
		id:               device.ID,
		signatureCounter: device.SignatureCounter,
		privateKey:       priv,
		algorithm:        device.Algorithm.String(),
		label:            device.Label,
	}

	return nil
}

// IncrementSignatureCounter increments the signature counter for a device in the persistence layer.
func (p *InMemory) IncrementSignatureCounter(ctx context.Context, id uuid.UUID) error {
	device, ok := p.storage[id]
	if !ok {
		return fmt.Errorf("device not found")
	}

	device.signatureCounter++

	return nil
}

// GetDevices returns all devices from the persistence layer.
func (p *InMemory) GetDevices(ctx context.Context) ([]domain.Device, error) {
	devices := make([]domain.Device, 0, len(p.storage))
	for _, device := range p.storage {
		kp, err := p.kpMarshaler.Unmarshal(domain.Algorithm(device.algorithm), device.privateKey)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal key pair: %w", err)
		}

		devices = append(devices, domain.Device{
			ID:               device.id,
			SignatureCounter: device.signatureCounter,
			KeyPair:          kp,
			Algorithm:        domain.Algorithm(device.algorithm),
			Label:            device.label,
		})
	}

	return devices, nil
}

// GetDevice returns a device from the persistence layer.
func (p *InMemory) GetDevice(ctx context.Context, id uuid.UUID) (domain.Device, error) {
	device, ok := p.storage[id]
	if !ok {
		return domain.Device{}, fmt.Errorf("device not found")
	}

	kp, err := p.kpMarshaler.Unmarshal(domain.Algorithm(device.algorithm), device.privateKey)
	if err != nil {
		return domain.Device{}, fmt.Errorf("could not unmarshal key pair: %w", err)
	}

	return domain.Device{
		ID:               device.id,
		SignatureCounter: device.signatureCounter,
		KeyPair:          kp,
		Algorithm:        domain.Algorithm(device.algorithm),
		Label:            device.label,
	}, nil
}

// SaveSignature saves a signature for a device in the persistence layer.
func (p *InMemory) SaveSignature(ctx context.Context, deviceID uuid.UUID, data domain.SignedData) error {
	device, ok := p.storage[deviceID]
	if !ok {
		return fmt.Errorf("device not found")
	}

	device.signatures = append(device.signatures, Signature{
		signature:    data.Signature,
		originalData: data.OriginalData,
		createdAt:    time.Now(),
	})

	return nil
}

// GetLastSignature returns the last signature for a device from the persistence layer.
func (p *InMemory) GetLastSignature(ctx context.Context, deviceID uuid.UUID) (domain.SignedData, error) {
	device, ok := p.storage[deviceID]
	if !ok {
		return domain.SignedData{}, fmt.Errorf("device not found")
	}

	if len(device.signatures) == 0 {
		return domain.SignedData{}, fmt.Errorf("no signatures found")
	}

	lastSignature := device.signatures[len(device.signatures)-1]
	return domain.SignedData{
		OriginalData: lastSignature.originalData,
		Signature:    lastSignature.signature,
	}, nil
}

// GetSignatures returns all signatures for a device from the persistence layer.
func (p *InMemory) GetSignatures(ctx context.Context, deviceID uuid.UUID) ([]domain.SignedData, error) {
	device, ok := p.storage[deviceID]
	if !ok {
		return nil, fmt.Errorf("device not found")
	}

	signatures := make([]domain.SignedData, 0, len(device.signatures))
	for _, signature := range device.signatures {
		signatures = append(signatures, domain.SignedData{
			OriginalData: signature.originalData,
			Signature:    signature.signature,
		})
	}

	return signatures, nil
}

// RunTransaction runs a transaction in the persistence layer. In this implementation it will get a mutex for specific
// device. I wrote this function like this to show how I would implement it with the regular database.
func (p *InMemory) RunTransaction(ctx context.Context, deviceID uuid.UUID, fn func(ctx context.Context) error) error {
	p.storage[deviceID].Lock()
	defer p.storage[deviceID].Unlock()

	return fn(ctx)
}
