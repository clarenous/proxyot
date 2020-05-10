package curve_test

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/clarenous/proxyot/curve"
)

func newRandPoint(typ curve.Curve) curve.Point {
	r := big.NewInt(rand.Int63())
	return curve.NewPoint(typ).ScalarBaseMult(r)
}

func newRandPointAndK(typ curve.Curve) (curve.Point, *big.Int) {
	r, k := big.NewInt(rand.Int63()), big.NewInt(rand.Int63())
	p := curve.NewPoint(typ).ScalarBaseMult(r)
	return p, k
}

func TestCurve_ScalarBaseMultTime(t *testing.T) {
	var round = 10_000
	testScalarBaseMultTime(round, curve.TypeG1, true)
	testScalarBaseMultTime(round, curve.TypeG2, true)
	testScalarBaseMultTime(round, curve.TypeGT, true)
}

func testScalarBaseMultTime(round int, typ curve.Curve, print bool) time.Duration {
	p, k := newRandPointAndK(typ)
	elapsed := pointScalarBaseMultTime(round, p, k)
	if print {
		fmt.Printf("curve: %s, round: %d, elapsed(ms): %d, avg(ms): %f\n",
			typ, round, elapsed.Milliseconds(), float64(elapsed.Milliseconds())/float64(round))
	}
	return elapsed
}

func pointScalarBaseMultTime(round int, point curve.Point, k *big.Int) time.Duration {
	start := time.Now()
	for i := 0; i < round; i++ {
		point.ScalarBaseMult(k)
	}
	return time.Since(start)
}

func TestCurve_ScalarMultTime(t *testing.T) {
	var round = 10_000
	testScalarMultTime(round, curve.TypeG1, true)
	testScalarMultTime(round, curve.TypeG2, true)
	testScalarMultTime(round, curve.TypeGT, true)
}

func testScalarMultTime(round int, typ curve.Curve, print bool) time.Duration {
	p, k := newRandPointAndK(typ)
	elapsed := pointScalarMultTime(round, p, k)
	if print {
		fmt.Printf("curve: %s, round: %d, elapsed(ms): %d, avg(ms): %f\n",
			typ, round, elapsed.Milliseconds(), float64(elapsed.Milliseconds())/float64(round))
	}
	return elapsed
}

func pointScalarMultTime(round int, point curve.Point, k *big.Int) time.Duration {
	start := time.Now()
	for i := 0; i < round; i++ {
		point.ScalarMult(point, k)
	}
	return time.Since(start)
}

func TestPairTime(t *testing.T) {
	var round = 100_000
	testPairTime(round, true)
}

func testPairTime(round int, print bool) time.Duration {
	// generate randoms g1 points
	g1Points := make([]*curve.G1, round)
	for i := range g1Points {
		g1Points[i] = newRandPoint(curve.TypeG1).(*curve.G1)
	}
	// generate randoms g2 points
	g2Points := make([]*curve.G2, round)
	for i := range g2Points {
		g2Points[i] = newRandPoint(curve.TypeG2).(*curve.G2)
	}
	// pairing
	start := time.Now()
	for i := 0; i < round; i++ {
		curve.Pair(g1Points[i], g2Points[i])
	}
	elapsed := time.Since(start)
	if print {
		fmt.Printf("curve: pairing, round: %d, elapsed(ms): %d, avg(ms): %f\n",
			round, elapsed.Milliseconds(), float64(elapsed.Milliseconds())/float64(round))
	}
	return elapsed
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
