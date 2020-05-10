package pre_test

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
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

func TestGenerateReKeyTime(t *testing.T) {
	var round = 10_000
	testGenerateReKeyTime(round, true)
}

func testGenerateReKeyTime(round int, print bool) time.Duration {
	// generate random filed numbers
	nums := make([]*big.Int, round*2)
	for i := range nums {
		n, err := curve.RandomFieldElement(rand.Reader)
		if err != nil {
			panic(err)
		}
		nums[i] = n
	}
	// generate re-keys
	start := time.Now()
	for i := 0; i < round; i++ {
		pre.GenerateReKey(nums[i], nums[i+round])
	}
	elapsed := time.Since(start)
	if print {
		fmt.Printf("pre: generate_re_key, round: %d, elapsed(ms): %d, avg(ms): %f\n",
			round, elapsed.Milliseconds(), float64(elapsed.Milliseconds())/float64(round))
	}
	return elapsed
}

func TestReEncryptTime(t *testing.T) {
	var round = 100_000
	testReEncryptTime(round, true)
}

func newRandPoint(typ curve.Curve) curve.Point {
	r := big.NewInt(mrand.Int63())
	return curve.NewPoint(typ).ScalarBaseMult(r)
}

func testReEncryptTime(round int, print bool) time.Duration {
	// generate randoms g1 points
	g1Points := make([]curve.Point, round)
	for i := range g1Points {
		g1Points[i] = newRandPoint(curve.TypeG1)
	}
	// generate randoms g2 points
	g2Points := make([]curve.Point, round)
	for i := range g2Points {
		g2Points[i] = newRandPoint(curve.TypeG2)
	}
	// pairing
	start := time.Now()
	for i := 0; i < round; i++ {
		pre.ReEncrypt(g1Points[i], g2Points[i])
	}
	elapsed := time.Since(start)
	if print {
		fmt.Printf("pre: re_encrypt, round: %d, elapsed(ms): %d, avg(ms): %f\n",
			round, elapsed.Milliseconds(), float64(elapsed.Milliseconds())/float64(round))
	}
	return elapsed
}

func init() {
	mrand.Seed(time.Now().UnixNano())
}
