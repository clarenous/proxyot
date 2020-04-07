package pre_test

import (
	"bytes"
	"crypto/rand"
	mrand "math/rand"
	"testing"
	"time"

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

	plaintext := make([]byte, 100)
	mrand.Seed(time.Now().UnixNano())
	if _, err = mrand.Read(plaintext); err != nil {
		t.Fatal(err)
	}

	// step 1: encrypt
	cipherBuf := bytes.NewBuffer(nil)
	A, err := pre.Encrypt(publicKeyA, pre.NewEncryptClosure(bytes.NewReader(plaintext), cipherBuf))
	if err != nil {
		t.Fatal(err)
	}
	ciphertext := cipherBuf.Bytes()

	// step 2: generate re-key
	rkAB := pre.GenerateReKey(a, b)

	// step 3: re-encrypt
	APrime := pre.ReEncrypt(A, rkAB)

	// step 4: decrypt by receiver
	receiverDeBuf := bytes.NewBuffer(nil)
	err = pre.DecryptByReceiver(APrime, b, pre.NewDecryptClosure(bytes.NewReader(ciphertext), receiverDeBuf))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plaintext, receiverDeBuf.Bytes()) {
		t.Errorf("decrypt by receiver error")
	}

	// step 5: decrypt by owner
	ownerDeBuf := bytes.NewBuffer(nil)
	err = pre.DecryptByOwner(A, a, pre.NewDecryptClosure(bytes.NewReader(ciphertext), ownerDeBuf))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plaintext, ownerDeBuf.Bytes()) {
		t.Errorf("decrypt by owner error")
	}
}
