package ot_test

import (
	"math/big"
	"testing"

	"github.com/clarenous/proxyot/ot"
)

func TestOT(t *testing.T) {
	var count int64 = 50
	var betaI int64 = 23
	var beta = big.NewInt(betaI)

	yp, r, err := ot.SealChoice(beta)
	if err != nil {
		t.Fatal(err)
	}

	kps, _, tg, err := ot.CalculateKeyPoints(yp, count)
	if err != nil {
		t.Fatal(err)
	}

	kpA := kps[betaI-1]
	kpB := ot.RevealKeyPoint(tg, r)
	if kpA.String() != kpB.String() {
		t.Errorf("Key Point not equal:\nAlice: %s\nBob: %s", kpA.String(), kpB.String())
	}
}
