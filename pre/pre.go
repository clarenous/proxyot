package pre

import (
	"crypto/rand"
	"math/big"

	"github.com/clarenous/proxyot/curve"
)

func Encrypt(publicKey curve.Point) (A curve.Point, B curve.Point, err error) {
	r, err := curve.RandomFieldElement(rand.Reader)
	if err != nil {
		return
	}
	// Ca = (A, B) = (r*PkA, rGt + Pm) =  (ra*G1, rGt + Pm)
	A = curve.NewPoint(publicKey.Curve()).ScalarMult(publicKey, r)
	B = newPairedPoint().ScalarBaseMult(r)
	return
}

func GenerateReKey(a, b *big.Int) (rkAB curve.Point) {
	// rkAB = a^-1 * pkB = (b/a) * G
	ia := new(big.Int).ModInverse(a, curve.Order) // a^-1 mod Order
	rkAB = newOtPoint().ScalarBaseMult(b)         // b * G
	rkAB = rkAB.ScalarMult(rkAB, ia)              // (b/a)*G
	return
}

func ReEncrypt(A, rkAB curve.Point) (APrime curve.Point) {
	APrime = curve.Pair(A.(*curve.G1), rkAB.(*curve.G2))
	return
}

func Decrypt(APrime curve.Point, b *big.Int) (B curve.Point) {
	ib := new(big.Int).ModInverse(b, curve.Order) // b^-1 mod Order
	B = newPairedPoint().ScalarMult(APrime, ib)   // B = rGt
	return
}

func newOtPoint() curve.Point {
	return curve.NewPoint(curve.TypeG2)
}

func newPairedPoint() curve.Point {
	return curve.NewPoint(curve.TypeGT)
}
