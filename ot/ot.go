package ot

import (
	"crypto/rand"
	"math/big"

	"github.com/clarenous/proxyot/curve"
)

// Params is the OT handler contains some parameters.
type Params struct{}

// New creates an OT handler.
func New() *Params {
	return &Params{}
}

// SealChoice refers to the process that receivers indicate the index of secret.
// The input argument beta is the index.
func SealChoice(beta *big.Int, pkA, pkB curve.Point) (Y curve.Point, L curve.Point, err error) {
	return defaultParams.SealChoice(beta, pkA, pkB)
}

// SealChoice refers to the process that receivers indicate the index of secret.
// The input argument beta is the index.
func (params *Params) SealChoice(beta *big.Int, pkA, pkB curve.Point) (Y curve.Point, L curve.Point, err error) {
	// random number l
	var l *big.Int
	l, err = curve.RandomFieldElement(rand.Reader)
	if err != nil {
		return
	}
	// calculate L = l * G
	L = newPoint().ScalarBaseMult(l)
	// calculate Y = beta * pkA + l * pkB
	Y = newPoint().ScalarMult(pkA, beta)
	Y.Add(Y, newPoint().ScalarMult(pkB, l))
	return
}

// CalculateKeyPoints refers to the process that sender creates some sealed key points
// given by the sender calculated Y point.
func CalculateKeyPoints(Y, L, pkA curve.Point, count int64) (kps []curve.Point, LPrime curve.Point, err error) {
	return defaultParams.CalculateKeyPoints(Y, L, pkA, count)
}

// CalculateKeyPoints refers to the process that sender creates some sealed key points
// given by the sender calculated Y point.
func (params *Params) CalculateKeyPoints(Y, L, pkA curve.Point, count int64) (kps []curve.Point, LPrime curve.Point, err error) {
	// random number t
	var t *big.Int
	t, err = curve.RandomFieldElement(rand.Reader)
	if err != nil {
		return
	}
	// calculate puzzles
	kps = make([]curve.Point, count)
	for i := int64(1); i <= count; i++ {
		kps[i-1] = params.calculateKeyPoint(t, Y, L, pkA, big.NewInt(i))
	}
	// calculate LPrime = t * L
	LPrime = newPoint().ScalarMult(L, t)
	return
}

func (params *Params) calculateKeyPoint(t *big.Int, yp, lp, pkA curve.Point, ordinal *big.Int) (kpi curve.Point) {
	// kpi = t * yp - i * t * pkA
	kpi = newPoint().ScalarMult(yp, t)
	temp := newPoint().ScalarMult(pkA, new(big.Int).Mul(ordinal, t))
	kpi.Add(kpi, temp.Neg(temp))
	return
}

// RevealKeyPoint refers to the process that receiver reveals real key point
// given by the sender calculated sealed key point.
func RevealKeyPoint(LPrime curve.Point, skB *big.Int) curve.Point {
	return defaultParams.RevealKeyPoint(LPrime, skB)
}

// RevealKeyPoint refers to the process that receiver reveals real key point
// given by the sender calculated sealed key point.
func (params *Params) RevealKeyPoint(LPrime curve.Point, skB *big.Int) curve.Point {
	return newPoint().ScalarMult(LPrime, skB)
}

func newPoint() curve.Point {
	return curve.NewPoint(curve.TypeG1)
}

var defaultParams *Params

func initDefaultParams() {
	defaultParams = &Params{}
}

func init() {
	initDefaultParams()
}
