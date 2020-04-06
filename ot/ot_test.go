package ot_test

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/clarenous/proxyot/curve"
	"github.com/clarenous/proxyot/ot"
)

func TestOT(t *testing.T) {
	var count int64 = 50
	var betaI int64 = 23
	var beta = big.NewInt(betaI)

	_, pkA, err := curve.NewRandomPoint(curve.TypeG1, rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	skB, pkB, err := curve.NewRandomPoint(curve.TypeG1, rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	Y, L, err := ot.SealChoice(beta, pkA, pkB)
	if err != nil {
		t.Fatal(err)
	}

	kps, LPrime, err := ot.CalculateKeyPoints(Y, L, pkA, count)
	if err != nil {
		t.Fatal(err)
	}

	kpA := kps[betaI-1]
	kpB := ot.RevealKeyPoint(LPrime, skB)
	if kpA.String() != kpB.String() {
		t.Errorf("Key Point not equal:\nAlice: %s\nBob: %s", kpA.String(), kpB.String())
	}
}
