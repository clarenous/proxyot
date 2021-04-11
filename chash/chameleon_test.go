package chash_test

import (
	"crypto/rand"
	"github.com/clarenous/proxyot/chash"
	"github.com/clarenous/proxyot/curve"
	"math/big"
	"testing"
	"time"
)

func TestVerify(t *testing.T) {
	m, _, _, _, Y, R, err := generateParams()
	if err != nil {
		t.Fatal(err)
	}
	ch := chash.ComputeHash(Y, R, m)
	if ok := chash.Verify(ch, Y, R, m); !ok {
		t.Error("verify failed")
	}
}

func TestComputeCollision(t *testing.T) {
	m, mp, x, r, Y, R, err := generateParams()
	if err != nil {
		t.Fatal(err)
	}
	oldH := chash.ComputeHash(Y, R, m)
	newH, _, _ := chash.ComputeCollision(Y, x, r, m, mp, curve.Order)
	if ok := oldH.Equals(newH); !ok {
		t.Error("collision failed")
	}
	t.Log("Old Hash:", oldH.String())
	t.Log("New Hash:", newH.String())
}

func TestComputeHashTime(t *testing.T) {
	var count = 1000
	// prepare
	m, _, _, _, Y, R, err := generateParams()
	if err != nil {
		t.Fatal(err)
	}
	// normal mode
	start := time.Now()
	for i := 0; i < count; i++ {
		chash.ComputeHash(Y, R, m)
	}
	spent := time.Since(start)
	t.Log(count, "rounds time spent:", spent.Seconds())
}

func TestComputeCollisionTime(t *testing.T) {
	var count = 1000
	// prepare
	m, _, x, r, Y, _, err := generateParams()
	if err != nil {
		t.Fatal(err)
	}
	mps := make([]*big.Int, count)
	for i := range mps {
		if mps[i], err = curve.RandomFieldElement(rand.Reader); err != nil {
			t.Fatal(err)
		}
	}
	// normal mode
	start := time.Now()
	for i := 0; i < count; i++ {
		chash.ComputeCollision(Y, x, r, m, mps[i], curve.Order)
	}
	spent := time.Since(start)
	t.Log(count, "rounds time spent:", spent.Seconds())
}

func generateParams() (m, mp, x, r *big.Int, Y, R curve.Point, err error) {
	if m, err = curve.RandomFieldElement(rand.Reader); err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	if mp, err = curve.RandomFieldElement(rand.Reader); err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	if x, Y, err = curve.NewRandomPoint(curve.TypeG1, rand.Reader); err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	if r, R, err = curve.NewRandomPoint(curve.TypeG2, rand.Reader); err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	return
}
