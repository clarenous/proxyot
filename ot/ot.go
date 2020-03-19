package ot

import (
	"crypto/rand"
	"math/big"

	"github.com/clarenous/proxyot/curve"
)

// Params is the OT handler contains some key parameters.
type Params struct {
	GeneratorH curve.Point
}

// New creates an OT handler with generator H.
func New(H curve.Point) *Params {
	return &Params{
		GeneratorH: H,
	}
}

// SealChoice refers to the process that receivers indicate the index of secret.
// The input argument beta is the index.
func SealChoice(beta *big.Int) (yp curve.Point, r *big.Int, err error) {
	return defaultParams.SealChoice(beta)
}

// SealChoice refers to the process that receivers indicate the index of secret.
// The input argument beta is the index.
func (params *Params) SealChoice(beta *big.Int) (yp curve.Point, r *big.Int, err error) {
	// random number r
	r, err = curve.RandomFieldElement(rand.Reader)
	if err != nil {
		return
	}
	// calculate R = g^r
	R := newPoint().ScalarBaseMult(r)
	// calculate B = h^beta
	B := newPoint().ScalarMult(params.GeneratorH, beta)
	// calculate yp = g^r * h^beta = R + B
	yp = newPoint().Add(R, B)
	return
}

// CalculateKeyPoints refers to the process that sender creates some sealed key points
// given by the sender calculated yp.
func CalculateKeyPoints(yp curve.Point, count int64) (kps []curve.Point, t *big.Int, tg curve.Point, err error) {
	return defaultParams.CalculateKeyPoints(yp, count)
}

// CalculateKeyPoints refers to the process that sender creates some sealed key points
// given by the sender calculated yp.
func (params *Params) CalculateKeyPoints(yp curve.Point, count int64) (kps []curve.Point, t *big.Int, T curve.Point, err error) {
	// random number t
	t, err = curve.RandomFieldElement(rand.Reader)
	if err != nil {
		return
	}
	// calculate T = g^t
	T = newPoint().ScalarBaseMult(t)
	// calculate puzzles
	kps = make([]curve.Point, count)
	for i := int64(1); i <= count; i++ {
		kps[i-1] = params.calculateKeyPoint(t, yp, big.NewInt(i))
	}
	return
}

func (params *Params) calculateKeyPoint(t *big.Int, yp curve.Point, ordinal *big.Int) curve.Point {
	// h^-i
	point := newPoint().ScalarMult(params.GeneratorH, big.NewInt(0).Sub(curve.Order, ordinal))
	// y * h^-i
	point = point.Add(yp, point)
	// (y * h^-i) ^ t
	return point.ScalarMult(point, t)
}

// RevealKeyPoint refers to the process that receiver reveals real key point
// given by the sender calculated sealed key point.
func RevealKeyPoint(T curve.Point, r *big.Int) curve.Point {
	return defaultParams.RevealKeyPoint(T, r)
}

// RevealKeyPoint refers to the process that receiver reveals real key point
// given by the sender calculated sealed key point.
func (params *Params) RevealKeyPoint(T curve.Point, r *big.Int) curve.Point {
	return newPoint().ScalarMult(T, r)
}

func newPoint() curve.Point {
	return curve.NewPoint(curve.TypeG1)
}

var defaultGeneratorH curve.Point

func initDefaultGeneratorH() error {
	h, err := curve.RandomFieldElement(rand.Reader)
	if err != nil {
		return err
	}
	defaultGeneratorH = newPoint().ScalarBaseMult(h)
	return nil
}

var defaultParams *Params

func initDefaultParams() {
	defaultParams = &Params{
		GeneratorH: defaultGeneratorH,
	}
}

func init() {
	if err := initDefaultGeneratorH(); err != nil {
		panic(err)
	}
	initDefaultParams()
}
