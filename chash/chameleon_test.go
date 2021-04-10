package chash_test

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/clarenous/proxyot/chash"
	"github.com/clarenous/proxyot/curve"
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
