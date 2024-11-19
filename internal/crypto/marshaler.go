package crypto

import (
	"fmt"
	"github.com/gren236/fiskaly-go-challenge/internal/domain"
)

type Marshaler struct {
	rsa *RSAMarshaler
	ecc *ECCMarshaler
}

func NewMarshaler() *Marshaler {
	return &Marshaler{
		rsa: NewRSAMarshaler(),
		ecc: NewECCMarshaler(),
	}
}

func (m Marshaler) Marshal(pair domain.KeyPair) ([]byte, []byte, error) {
	switch kp := pair.(type) {
	case *RSAKeyPair:
		return m.rsa.Marshal(*kp)
	case *ECCKeyPair:
		return m.ecc.Marshal(*kp)
	default:
		return nil, nil, fmt.Errorf("unsupported key pair type")
	}
}

func (m Marshaler) Unmarshal(algo domain.Algorithm, privateKeyBytes []byte) (domain.KeyPair, error) {
	switch algo {
	case domain.AlgorithmRSA:
		return m.rsa.Unmarshal(privateKeyBytes)
	case domain.AlgorithmECC:
		return m.ecc.Unmarshal(privateKeyBytes)
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algo)
	}
}
