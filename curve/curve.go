package curve

import (
	"crypto/rand"
	"io"
	"math/big"

	"github.com/cloudflare/bn256"
)

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
