package curve

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"math/big"

	"github.com/cloudflare/bn256"
)

const (
	// TypeG1 represents point type G1
	TypeG1 Curve = iota
	// TypeG2 represents point type G2
	TypeG2
	// TypeGT represents point type GT
	TypeGT
)

type Curve uint32

func (curve Curve) String() (s string) {
	switch curve {
	case TypeG1:
		s = "bn256.G1"
	case TypeG2:
		s = "bn256.G2"
	case TypeGT:
		s = "bn256.GT"
	default:
		s = fmt.Sprintf("invalid(%d)", curve)
	}
	return
}

var (
	Order = bn256.Order
)

// RandomFieldElement returns x where x is a random, non-zero number read from r.
func RandomFieldElement(r io.Reader) (*big.Int, error) {
	var k *big.Int
	var err error

	for {
		k, err = rand.Int(r, Order)
		if err != nil {
			return nil, err
		}
		if k.Sign() > 0 {
			break
		}
	}

	return k, nil
}

// DeriveFieldElementFromPoint derives a field element k by the following rules:
// (1) Calculate h = SHA256(point.Marshal())
// (2) Calculate k = BigInt(h) % Order
// (3) If k == 0, calculate h = SHA256(h), then jump to (2)
func DeriveFieldElementFromPoint(point Point) *big.Int {
	var h = sha256.Sum256(point.Marshal())
	var k = new(big.Int)
	for {
		k.SetBytes(h[:])
		k.Mod(k, Order)
		if k.Sign() > 0 {
			break
		}
		h = sha256.Sum256(h[:])
	}
	return k
}
