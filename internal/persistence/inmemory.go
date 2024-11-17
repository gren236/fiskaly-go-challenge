package persistence

import (
	"context"
	"fmt"
	"github.com/fiskaly/coding-challenges/signing-service-challenge/internal/domain"
	"github.com/google/uuid"
	"sync"
	"time"
)

type Signature struct {
	domain.SignedData // We probably do not need a whole signature result here, just the signature, but for simplicity we will use the whole struct

	createdAt time.Time
}

type Device struct {
	domain.Device
	signatures []Signature

	sync.Mutex // We need to lock the device when adding a new signature
}

// InMemory is an in-memory implementation of the persistence layer.
type InMemory struct {
	storage map[uuid.UUID]*Device
}

// NewInMemory creates a new InMemory persistence layer.
func NewInMemory() *InMemory {
	return &InMemory{
		storage: make(map[uuid.UUID]*Device),
	}
}

// CreateDevice creates a new device in the persistence layer.
func (p *InMemory) CreateDevice(ctx context.Context, device domain.Device) error {
	p.storage[device.ID] = &Device{
		Device: device,
	}

	return nil
}

// GetDevices returns all devices from the persistence layer.
func (p *InMemory) GetDevices(ctx context.Context) ([]domain.Device, error) {
	devices := make([]domain.Device, 0, len(p.storage))
	for _, device := range p.storage {
		devices = append(devices, device.Device)
	}

	return devices, nil
}

// GetDevice returns a device from the persistence layer.
func (p *InMemory) GetDevice(ctx context.Context, id uuid.UUID) (domain.Device, error) {
	device, ok := p.storage[id]
	if !ok {
		return domain.Device{}, fmt.Errorf("device not found")
	}

	return device.Device, nil
}

// SaveSignature saves a signature for a device in the persistence layer.
func (p *InMemory) SaveSignature(ctx context.Context, deviceID uuid.UUID, data domain.SignedData) error {
	device, ok := p.storage[deviceID]
	if !ok {
		return fmt.Errorf("device not found")
	}

	device.signatures = append(device.signatures, Signature{
		SignedData: data,
		createdAt:  time.Now(),
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

	// return the last signature based on the timestamp
	lastSignature := device.signatures[0]
	for _, signature := range device.signatures {
		if signature.createdAt.After(lastSignature.createdAt) {
			lastSignature = signature
		}
	}

	return lastSignature.SignedData, nil
}

// GetSignatures returns all signatures for a device from the persistence layer.
func (p *InMemory) GetSignatures(ctx context.Context, deviceID uuid.UUID) ([]domain.SignedData, error) {
	device, ok := p.storage[deviceID]
	if !ok {
		return nil, fmt.Errorf("device not found")
	}

	signatures := make([]domain.SignedData, 0, len(device.signatures))
	for _, signature := range device.signatures {
		signatures = append(signatures, signature.SignedData)
	}

	return signatures, nil
}

// RunTransaction runs a transaction in the persistence layer.
func (p *InMemory) RunTransaction(ctx context.Context, deviceID uuid.UUID, fn func(ctx context.Context) error) error {
	p.storage[deviceID].Lock()
	defer p.storage[deviceID].Unlock()

	return fn(ctx)
}
