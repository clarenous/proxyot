package curve

import (
	"math/big"

	"github.com/cloudflare/bn256"
)

const (
	// TypeG1 represents point type G1
	TypeG1 = iota
	// TypeG2 represents point type G2
	TypeG2
	// TypeGT represents point type GT
	TypeGT
)

type Point interface {
	Name() string
	String() string
	Marshal() []byte
	Unmarshal(m []byte) ([]byte, error)
	ScalarBaseMult(k *big.Int) Point
	ScalarMult(a Point, k *big.Int) Point
	Add(a, b Point) Point
	Neg(a Point) Point
	Set(a Point) Point
}

func NewPoint(typ int) (point Point) {
	switch typ {
	case TypeG1:
		point = new(G1)
	case TypeG2:
		point = new(G2)
	case TypeGT:
		point = new(GT)
	}
	return
}

type G1 bn256.G1

func (e *G1) Name() string {
	return "bn256.G1"
}

func (e *G1) String() string {
	return (*bn256.G1)(e).String()
}

func (e *G1) Marshal() []byte {
	return (*bn256.G1)(e).Marshal()
}

func (e *G1) Unmarshal(m []byte) ([]byte, error) {
	return (*bn256.G1)(e).Unmarshal(m)
}

func (e *G1) ScalarBaseMult(k *big.Int) Point {
	(*bn256.G1)(e).ScalarBaseMult(k)
	return e
}

func (e *G1) ScalarMult(a Point, k *big.Int) Point {
	(*bn256.G1)(e).ScalarMult((*bn256.G1)(a.(*G1)), k)
	return e
}

func (e *G1) Add(a, b Point) Point {
	(*bn256.G1)(e).Add((*bn256.G1)(a.(*G1)), (*bn256.G1)(b.(*G1)))
	return e
}

func (e *G1) Neg(a Point) Point {
	(*bn256.G1)(e).Neg((*bn256.G1)(a.(*G1)))
	return e
}

func (e *G1) Set(a Point) Point {
	(*bn256.G1)(e).Set((*bn256.G1)(a.(*G1)))
	return e
}

type G2 bn256.G2

func (e *G2) Name() string {
	return "bn256.G2"
}

func (e *G2) String() string {
	return (*bn256.G2)(e).String()
}

func (e *G2) Marshal() []byte {
	return (*bn256.G2)(e).Marshal()
}

func (e *G2) Unmarshal(m []byte) ([]byte, error) {
	return (*bn256.G2)(e).Unmarshal(m)
}

func (e *G2) ScalarBaseMult(k *big.Int) Point {
	(*bn256.G2)(e).ScalarBaseMult(k)
	return e
}

func (e *G2) ScalarMult(a Point, k *big.Int) Point {
	(*bn256.G2)(e).ScalarMult((*bn256.G2)(a.(*G2)), k)
	return e
}

func (e *G2) Add(a, b Point) Point {
	(*bn256.G2)(e).Add((*bn256.G2)(a.(*G2)), (*bn256.G2)(b.(*G2)))
	return e
}

func (e *G2) Neg(a Point) Point {
	(*bn256.G2)(e).Neg((*bn256.G2)(a.(*G2)))
	return e
}

func (e *G2) Set(a Point) Point {
	(*bn256.G2)(e).Set((*bn256.G2)(a.(*G2)))
	return e
}

type GT bn256.GT

func Pair(g1 *G1, g2 *G2) *GT {
	gt := bn256.Pair((*bn256.G1)(g1), (*bn256.G2)(g2))
	return (*GT)(gt)
}

func (e *GT) Name() string {
	return "bn256.GT"
}

func (e *GT) String() string {
	return (*bn256.GT)(e).String()
}

func (e *GT) Marshal() []byte {
	return (*bn256.GT)(e).Marshal()
}

func (e *GT) Unmarshal(m []byte) ([]byte, error) {
	return (*bn256.GT)(e).Unmarshal(m)
}

func (e *GT) ScalarBaseMult(k *big.Int) Point {
	(*bn256.GT)(e).ScalarBaseMult(k)
	return e
}

func (e *GT) ScalarMult(a Point, k *big.Int) Point {
	(*bn256.GT)(e).ScalarMult((*bn256.GT)(a.(*GT)), k)
	return e
}

func (e *GT) Add(a, b Point) Point {
	(*bn256.GT)(e).Add((*bn256.GT)(a.(*GT)), (*bn256.GT)(b.(*GT)))
	return e
}

func (e *GT) Neg(a Point) Point {
	(*bn256.GT)(e).Neg((*bn256.GT)(a.(*GT)))
	return e
}

func (e *GT) Set(a Point) Point {
	(*bn256.GT)(e).Set((*bn256.GT)(a.(*GT)))
	return e
}
