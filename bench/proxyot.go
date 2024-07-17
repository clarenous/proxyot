package bench

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"testing"

	"github.com/clarenous/proxyot/curve"
	"github.com/clarenous/proxyot/ot"
	"github.com/clarenous/proxyot/pre"
)

func benchScalarBaseMult(b *testing.B, typ curve.Curve) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		p, k := randPoint(typ), randFiledElement()
		b.StartTimer()
		p.ScalarBaseMult(k)
	}
}

func benchScalarMult(b *testing.B, typ curve.Curve) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		p, k := randPoint(typ), randFiledElement()
		b.StartTimer()
		p.ScalarMult(p, k)
	}
}

func benchScalarDiv(b *testing.B, typ curve.Curve) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		p, k := randPoint(typ), randFiledElement()
		b.StartTimer()
		ik := new(big.Int).ModInverse(k, curve.Order)
		p.ScalarMult(p, ik)
	}
}

func benchPointAdd(b *testing.B, typ curve.Curve) {
	p1, p2 := randPoint(typ), randPoint(typ)
	for i := 0; i < b.N; i++ {
		p1.Add(p1, p2)
	}
}

func benchPair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		p1 := randPoint(curve.TypeG1).(*curve.G1)
		p2 := randPoint(curve.TypeG2).(*curve.G2)
		b.StartTimer()
		curve.Pair(p1, p2)
	}
}

func benchGenerateKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		t := randFiledElement()
		y := randPoint(curve.TypeG1)
		pkA := randPoint(curve.TypeG1)
		ordinal := big.NewInt(rand.Int63n(100_000))
		b.StartTimer()
		kpi := calculateKeyPoint(t, y, pkA, ordinal)
		curve.DeriveFieldElementFromPoint(kpi)
	}
}

func benchGenerateReKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		x, y := randFiledElement(), randFiledElement()
		b.StartTimer()
		pre.GenerateReKey(x, y)
	}
}

func benchReEncrypt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		p1, p2 := randPoint(curve.TypeG1), randPoint(curve.TypeG2)
		b.StartTimer()
		pre.ReEncrypt(p1, p2)
	}
}

func benchEncrypt(b *testing.B, size int64) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		data := mustMakeRandomBytes(size)
		closure := pre.NewEncryptClosure(bytes.NewReader(data), ioutil.Discard)
		pk := randPoint(curve.TypeG1)
		b.StartTimer()
		if _, err := pre.Encrypt(pk, closure); err != nil {
			b.Fatal(err)
		}
	}
}

func benchAESEncrypt(b *testing.B, size int64) {
	b.StopTimer()
	data := mustMakeRandomBytes(size)
	key := randPoint(curve.TypeGT).Marshal()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if err := pre.NewEncryptClosure(bytes.NewReader(data), ioutil.Discard)(key); err != nil {
			b.Fatal(err)
		}
	}
}

func benchDecrypt(b *testing.B, size int64) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// generate key pair
		skA, pkA := randPointAndKey(curve.TypeG1)
		skB := randFiledElement()
		// generate random data and encrypt
		data := mustMakeRandomBytes(size)
		cipherBuf := bytes.NewBuffer(nil)
		encClosure := pre.NewEncryptClosure(bytes.NewReader(data), cipherBuf)
		A, err := pre.Encrypt(pkA, encClosure)
		if err != nil {
			b.Fatal(err)
		}
		encrypted := cipherBuf.Bytes()
		// re_encrypt
		rkAB := pre.GenerateReKey(skA, skB)
		APrime := pre.ReEncrypt(A, rkAB)
		// decrypt closure
		decClosure := pre.NewDecryptClosure(bytes.NewReader(encrypted), ioutil.Discard)

		b.StartTimer()
		if err := pre.DecryptByReceiver(APrime, skB, decClosure); err != nil {
			b.Fatal(err)
		}
	}
}

func benchAESDecrypt(b *testing.B, size int64) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		data := mustMakeRandomBytes(size)
		cipherBuf := bytes.NewBuffer(make([]byte, 0, size+4096))
		encryptClosure := pre.NewEncryptClosure(bytes.NewReader(data), cipherBuf)
		key := randPoint(curve.TypeGT).Marshal()
		if err := encryptClosure(key); err != nil {
			b.Fatal(err)
		}
		decryptClosure := pre.NewDecryptClosure(bytes.NewReader(cipherBuf.Bytes()), ioutil.Discard)
		b.StartTimer()
		if err := decryptClosure(key); err != nil {
			b.Fatal(err)
		}
	}
}

func benchShareMessage(b *testing.B, size int64, count int64) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// generate key pair
		skA, pkA := randPointAndKey(curve.TypeG1)
		skB, pkB := randPointAndKey(curve.TypeG1)
		// random target
		target := rand.Int63n(count) + 1
		// generate random messages and encrypt
		As, ciphertext, err := initShareMessages(pkA, size, count, target)
		if err != nil {
			b.Fatal(err)
		}

		b.StartTimer()
		if err = shareMessages(skA, skB, pkA, pkB, As, ciphertext, count, target, false); err != nil {
			b.Fatal(err)
		}
	}
}

func randPoint(typ curve.Curve) curve.Point {
	return curve.NewPoint(typ).ScalarBaseMult(randFiledElement())
}

func randPointAndKey(typ curve.Curve) (*big.Int, curve.Point) {
	k := randFiledElement()
	p := curve.NewPoint(typ).ScalarBaseMult(k)
	return k, p
}

func randFiledElement() *big.Int {
	var k = new(big.Int)
	var buf [32]byte
	for {
		binary.BigEndian.PutUint64(buf[:], rand.Uint64())
		binary.BigEndian.PutUint64(buf[8:], rand.Uint64())
		binary.BigEndian.PutUint64(buf[16:], rand.Uint64())
		binary.BigEndian.PutUint64(buf[24:], rand.Uint64())
		k.SetBytes(buf[:])
		k.Mod(k, curve.Order)
		if k.Sign() > 0 {
			break
		}
	}
	return k
}

func calculateKeyPoint(t *big.Int, yp, pkA curve.Point, ordinal *big.Int) (kpi curve.Point) {
	// kpi = t * yp - i * t * pkA
	kpi = curve.NewPoint(curve.TypeG1).ScalarMult(yp, t)
	temp := curve.NewPoint(curve.TypeG1).ScalarMult(pkA, new(big.Int).Mul(ordinal, t))
	kpi.Add(kpi, temp.Neg(temp))
	return
}

func deriveKeysFromPoints(points []curve.Point) []*big.Int {
	keys := make([]*big.Int, len(points))
	for i := range points {
		keys[i] = curve.DeriveFieldElementFromPoint(points[i])
	}
	return keys
}

func mustReadRandomBytes(buf []byte) {
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
}

func mustMakeRandomBytes(size int64) []byte {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return buf
}

func mustMakeNRandomBytes(size int64, n int) [][]byte {
	result := make([][]byte, n)
	for i := range result {
		result[i] = mustMakeRandomBytes(size)
	}
	return result
}

func encryptMessages(pkA curve.Point, messages [][]byte) (As []curve.Point, ciphertexts [][]byte, err error) {
	for i := range messages {
		a, ciphertext, err := encryptMessage(pkA, messages[i])
		if err != nil {
			return nil, nil, fmt.Errorf("encryptMessages: %w", err)
		}
		As = append(As, a)
		ciphertexts = append(ciphertexts, ciphertext)
	}
	return
}

func encryptMessage(pkA curve.Point, message []byte) (A curve.Point, ciphertext []byte, err error) {
	cipherBuf := bytes.NewBuffer(nil)
	A, err = pre.Encrypt(pkA, pre.NewEncryptClosure(bytes.NewReader(message), cipherBuf))
	if err != nil {
		return nil, nil, fmt.Errorf("alice encryptMessage: %w", err)
	}
	ciphertext = cipherBuf.Bytes()
	return
}

func decryptMessage(skB *big.Int, APrime, LPrime curve.Point, cipher []byte) (message []byte, err error) {
	kp := ot.RevealKeyPoint(LPrime, skB)
	b := curve.DeriveFieldElementFromPoint(kp)
	buf := bytes.NewBuffer(nil)
	err = pre.DecryptByReceiver(APrime, b, pre.NewDecryptClosure(bytes.NewReader(cipher), buf))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func initShareMessages(pkA curve.Point, size, n, target int64) (As []curve.Point, ciphertext []byte, err error) {
	As = make([]curve.Point, n)
	// save memory
	for i := int64(0); i < n; i++ {
		if i == target-1 {
			msg := make([]byte, size)
			mustReadRandomBytes(msg)
			if As[i], ciphertext, err = encryptMessage(pkA, msg); err != nil {
				return nil, nil, err
			}
			msg = nil
		} else {
			As[i] = curve.NewPoint(pkA.Curve()).ScalarMult(pkA, randFiledElement())
		}
	}
	return
}

func shareMessages(skA, skB *big.Int, pkA, pkB curve.Point, As []curve.Point, ciphertext []byte, count, target int64, execDecrypt bool) error {
	// bob seal choice
	Y, L, err := ot.SealChoice(big.NewInt(target), pkA, pkB)
	if err != nil {
		return err
	}
	// alice calculate key points
	kps, LPrime, err := ot.CalculateKeyPoints(Y, L, pkA, count)
	if err != nil {
		return err
	}
	// alice derive bob decrypt key
	bobKeys := deriveKeysFromPoints(kps)
	// alice generate re_key
	reKeys := make([]curve.Point, len(bobKeys))
	for i := range bobKeys {
		reKeys[i] = pre.GenerateReKey(skA, bobKeys[i])
	}
	// proxy re_encrypt
	APrimes := make([]curve.Point, len(As))
	for i := range reKeys {
		APrimes[i] = pre.ReEncrypt(As[i], reKeys[i])
	}
	// bob decrypt
	if execDecrypt {
		kp := ot.RevealKeyPoint(LPrime, skB)
		bobKey := curve.DeriveFieldElementFromPoint(kp)
		return pre.DecryptByReceiver(APrimes[target-1], bobKey, pre.NewDecryptClosure(bytes.NewReader(ciphertext), ioutil.Discard))
	}
	return nil
}
