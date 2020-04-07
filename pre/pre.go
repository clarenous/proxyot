package pre

import (
	"crypto/rand"
	"math/big"

	"github.com/clarenous/proxyot/curve"
)

func Encrypt(publicKey curve.Point, encryptFunc EncryptClosure) (A curve.Point, err error) {
	r, err := curve.RandomFieldElement(rand.Reader)
	if err != nil {
		return
	}
	// Ca = (A, B) = (r*PkA, rGt + Pm) =  (ra*G, rGt + Pm)
	A = curve.NewPoint(publicKey.Curve()).ScalarMult(publicKey, r)
	B := newPairedPoint().ScalarBaseMult(r)
	err = encryptFunc(B.Marshal())
	return
}

func GenerateReKey(a, b *big.Int) (rkAB curve.Point) {
	// rkAB = a^-1 * pkB = (b/a) * G
	ia := new(big.Int).ModInverse(a, curve.Order) // a^-1 mod Order
	rkAB = newTwistPoint().ScalarBaseMult(b)      // b * G
	rkAB = rkAB.ScalarMult(rkAB, ia)              // (b/a)*G
	return
}

func ReEncrypt(A, rkAB curve.Point) (APrime curve.Point) {
	APrime = curve.Pair(A.(*curve.G1), rkAB.(*curve.G2))
	return
}

func DecryptByReceiver(APrime curve.Point, b *big.Int, decryptFunc DecryptClosure) (err error) {
	ib := new(big.Int).ModInverse(b, curve.Order) // b^-1 mod Order
	B := newPairedPoint().ScalarMult(APrime, ib)  // B = rGt
	err = decryptFunc(B.Marshal())
	return
}

func DecryptByOwner(A curve.Point, a *big.Int, decryptFunc DecryptClosure) (err error) {
	ia := new(big.Int).ModInverse(a, curve.Order) // b^-1 mod Order
	rG := newPoint().ScalarMult(A, ia)
	B := curve.Pair(rG.(*curve.G1), oneTwistPoint.(*curve.G2)) // B = rGt
	err = decryptFunc(B.Marshal())
	return
}

var oneTwistPoint = newTwistPoint().ScalarBaseMult(big.NewInt(1))

func newPoint() curve.Point {
	return curve.NewPoint(curve.TypeG1)
}

func newTwistPoint() curve.Point {
	return curve.NewPoint(curve.TypeG2)
}

func newPairedPoint() curve.Point {
	return curve.NewPoint(curve.TypeGT)
}
