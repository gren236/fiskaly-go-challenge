package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/gren236/fiskaly-go-challenge/internal/domain"
)

type Generator struct {
	rsa *RSAGenerator
	ecc *ECCGenerator
}

// NewGenerator creates a new Generator.
func NewGenerator() *Generator {
	return &Generator{
		rsa: &RSAGenerator{},
		ecc: &ECCGenerator{},
	}
}

func (g *Generator) GenerateKeyPair(algorithm domain.Algorithm) (domain.KeyPair, error) {
	switch algorithm {
	case domain.AlgorithmECC:
		return g.ecc.Generate()
	case domain.AlgorithmRSA:
		return g.rsa.Generate()
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// RSAGenerator generates an RSA key pair.
type RSAGenerator struct{}

// Generate generates a new RSAKeyPair.
func (g *RSAGenerator) Generate() (*RSAKeyPair, error) {
	// Security has been ignored for the sake of simplicity.
	key, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		return nil, err
	}

	return &RSAKeyPair{
		Public:  &key.PublicKey,
		Private: key,
	}, nil
}

// ECCGenerator generates an ECC key pair.
type ECCGenerator struct{}

// Generate generates a new ECCKeyPair.
func (g *ECCGenerator) Generate() (*ECCKeyPair, error) {
	// Security has been ignored for the sake of simplicity.
	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return &ECCKeyPair{
		Public:  &key.PublicKey,
		Private: key,
	}, nil
}
