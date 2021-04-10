package chash

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"

	"github.com/clarenous/proxyot/curve"
)

const Size = sha256.Size

type ChameleonHash [Size]byte

func (ch ChameleonHash) Equals(target ChameleonHash) bool {
	return ch == target
}

func (ch ChameleonHash) String() string {
	return hex.EncodeToString(ch[:])
}

// ComputeHash computes the chameleon hash.
func ComputeHash(Y, R curve.Point, m *big.Int) ChameleonHash {
	Qm := curve.NewPoint(curve.TypeGT).ScalarBaseMult(m)
	pairedYR := curve.Pair(Y.(*curve.G1), R.(*curve.G2))
	h := curve.NewPoint(curve.TypeGT).Add(Qm, pairedYR)
	return sha256.Sum256(h.Marshal())
}

// Verify verifies a given target with original params.
func Verify(target ChameleonHash, Y, R curve.Point, m *big.Int) bool {
	computed := ComputeHash(Y, R, m)
	return computed.Equals(target)
}

// ComputeCollision computes an instance of collision of the given chameleon hash.
func ComputeCollision(Y curve.Point, x, r, m, mp, q *big.Int) (ChameleonHash, *big.Int, curve.Point) {
	// compute rp such that m + x * r = mp + x * rp (mod q)
	temp := new(big.Int)
	temp.Mul(x, r).Add(temp, m).Sub(temp, mp).Mod(temp, q) // m + x * r - mp = x * rp (mod q)
	ix := new(big.Int).ModInverse(x, q)
	rp := temp.Mul(temp, ix)
	Rp := curve.NewPoint(curve.TypeG2).ScalarBaseMult(rp)
	ch := ComputeHash(Y, Rp, mp)
	return ch, rp, Rp
}
