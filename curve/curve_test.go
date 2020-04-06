package curve_test

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/clarenous/proxyot/curve"
)

func TestScalarBaseMult(t *testing.T) {
	var round = 10000
	var randoms = make([]*big.Int, round)

	var collectRandoms = func() {
		var err error
		for i := range randoms {
			if randoms[i], err = curve.RandomFieldElement(rand.Reader); err != nil {
				t.Fatal(i, "collect randoms", err)
			}
		}
	}

	var testFunc = func(typ curve.Curve) time.Duration {
		start := time.Now()
		for i := range randoms {
			curve.NewPoint(typ).ScalarBaseMult(randoms[i])
		}
		elapsed := time.Since(start)
		return elapsed
	}

	for _, typ := range []curve.Curve{curve.TypeG1, curve.TypeG2, curve.TypeGT} {
		collectRandoms()
		usage := testFunc(typ)
		fmt.Println(typ, "time used:", usage.Seconds())
	}
}

func TestPairing(t *testing.T) {
	var round = 10000
	var randoms = make([]*big.Int, round)

	var collectRandoms = func() {
		var err error
		for i := range randoms {
			if randoms[i], err = curve.RandomFieldElement(rand.Reader); err != nil {
				t.Fatal(i, "collect randoms", err)
			}
		}
	}

	var baseMultFunc = func(typ curve.Curve, result []curve.Point) time.Duration {
		start := time.Now()
		for i := range randoms {
			result[i] = curve.NewPoint(typ).ScalarBaseMult(randoms[i])
		}
		elapsed := time.Since(start)
		return elapsed
	}

	var pairingFunc = func(g1s, g2s []curve.Point) time.Duration {
		start := time.Now()
		for i := range g1s {
			curve.Pair(g1s[i].(*curve.G1), g2s[i].(*curve.G2))
		}
		elapsed := time.Since(start)
		return elapsed
	}

	collectRandoms()
	g1s := make([]curve.Point, round)
	usage1 := baseMultFunc(curve.TypeG1, g1s)
	fmt.Println(curve.TypeG1, "scalar base mult time used:", usage1.Seconds())

	collectRandoms()
	g2s := make([]curve.Point, round)
	usage2 := baseMultFunc(curve.TypeG2, g2s)
	fmt.Println(curve.TypeG2, "scalar base mult time used:", usage2.Seconds())

	usageT := pairingFunc(g1s, g2s)
	fmt.Println("pairing time used:", usageT.Seconds())
}
