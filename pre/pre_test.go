package pre_test

import (
	"crypto/rand"
	"testing"

	"github.com/clarenous/proxyot/curve"
	"github.com/clarenous/proxyot/pre"
)

func TestPre(t *testing.T) {
	a, publicKeyA, err := curve.NewRandomPoint(curve.TypeG1, rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	b, err := curve.RandomFieldElement(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	// step 1: encrypt
	A, B, err := pre.Encrypt(publicKeyA)
	if err != nil {
		t.Fatal(err)
	}

	// step 2: generate re-key
	rkAB := pre.GenerateReKey(a, b)

	// step 3: re-encrypt
	APrime := pre.ReEncrypt(A, rkAB)

	// step 4: decrypt
	recoveredB := pre.Decrypt(APrime, b)

	// verify
	if B.String() != recoveredB.String() {
		t.Errorf("Key Point not equal:\nB: %s\nrecoveredB: %s", B.String(), recoveredB.String())
	}
}
