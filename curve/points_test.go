package curve_test

import (
	"encoding/hex"
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

func TestDistributiveLaw(t *testing.T) {
	// point A, B (on G1)
	pA, pB := newRandPoint(curve.TypeG1), newRandPoint(curve.TypeG1)

	// point C (on G2)
	pC := newRandPoint(curve.TypeG2)

	// calculate left side
	leftG1Point := curve.NewPoint(curve.TypeG1)
	leftG1Point.Add(pA, pB)
	leftGtPoint := curve.Pair(leftG1Point.(*curve.G1), pC.(*curve.G2))

	// calculate right side
	rightGtPoint1 := curve.Pair(pA.(*curve.G1), pC.(*curve.G2))
	rightGtPoint2 := curve.Pair(pB.(*curve.G1), pC.(*curve.G2))
	rightGtPoint := curve.NewPoint(curve.TypeGT).Add(rightGtPoint1, rightGtPoint2)

	// check equality between left and right side
	leftGtPointStr := leftGtPoint.String()
	rightGtPointStr := rightGtPoint.String()
	fmt.Println("Left:", leftGtPointStr)
	fmt.Println("Right:", rightGtPointStr)
	if leftGtPointStr != rightGtPointStr {
		t.Error("Failed on TestDistributiveLaw")
	}
}

func TestPrintOut(t *testing.T) {
	// aG1, aG2, aGT
	a := big.NewInt(rand.Int63())
	aG1 := curve.NewPoint(curve.TypeG1).ScalarBaseMult(a)
	aG2 := curve.NewPoint(curve.TypeG2).ScalarBaseMult(a)
	aGT := curve.NewPoint(curve.TypeGT).ScalarBaseMult(a)

	// print out
	fmt.Println("aG1:", hex.EncodeToString(aG1.Marshal()))
	fmt.Println("aG2:", hex.EncodeToString(aG2.Marshal()))
	fmt.Println("aGT:", hex.EncodeToString(aGT.Marshal()))
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
