package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/fiskaly/coding-challenges/signing-service-challenge/internal/domain"
)

type SignerCreator struct{}

func NewSignerCreator() *SignerCreator {
	return &SignerCreator{}
}

// CreateSigner creates a new signer. Usually, it's not idiomatic in Go to return an interface instead of concrete type.
// However, in this case, it's necessary to return an interface because the concrete type of the signer is determined
// at runtime. BTW, Go std libraries use this pattern in some places, e.g., gob package.
func (sc *SignerCreator) CreateSigner(kp domain.KeyPair) (domain.Signer, error) {
	switch kp := kp.(type) {
	case ECCKeyPair:
		return &ECCSigner{keyPair: kp}, nil
	case RSAKeyPair:
		return &RSASigner{keyPair: kp}, nil
	default:
		return nil, fmt.Errorf("unsupported key pair type")
	}
}

// ECCSigner is a signer implementation for ECC key pairs.
type ECCSigner struct {
	keyPair ECCKeyPair
}

func (es *ECCSigner) Sign(dataToBeSigned []byte) ([]byte, error) {
	rawDataHash, err := hashData(dataToBeSigned)
	if err != nil {
		return nil, err
	}

	data, err := es.keyPair.Private.Sign(rand.Reader, rawDataHash, crypto.SHA256)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	return data, nil
}

// RSASigner is a signer implementation for RSA key pairs.
type RSASigner struct {
	keyPair RSAKeyPair
}

func (rs *RSASigner) Sign(dataToBeSigned []byte) ([]byte, error) {
	rawDataHash, err := hashData(dataToBeSigned)
	if err != nil {
		return nil, err
	}

	data, err := rs.keyPair.Private.Sign(rand.Reader, rawDataHash, crypto.SHA256)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	return data, nil
}

func hashData(data []byte) ([]byte, error) {
	sha256Hash := sha256.New()
	_, err := sha256Hash.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to hash data: %w", err)
	}

	return sha256Hash.Sum(nil), nil
}
