package node_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	mrand "math/rand"
	"sync"
	"testing"
	"time"

	"github.com/clarenous/proxyot/curve"
	"github.com/clarenous/proxyot/node"
	"github.com/clarenous/proxyot/node/protocol"
	"github.com/clarenous/proxyot/ot"
	"github.com/clarenous/proxyot/pre"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/multiformats/go-multihash"
)

func TestNode(t *testing.T) {
	node1 := node.NewRandomBaseNode(41001)
	log.Println("running node1...", node1.Host.ID(), node1.Host.Addrs(), node1.Host.Peerstore().Peers())
	defer node1.Host.Close()

	node2 := node.NewRandomBaseNode(41002)
	log.Println("running node2...", node2.Host.ID(), node2.Host.Addrs(), node2.Host.Peerstore().Peers())
	defer node2.Host.Close()

	node1.Host.Peerstore().AddAddrs(node2.Host.ID(), node2.Host.Addrs(), peerstore.PermanentAddrTTL)
	//node2.Host.Peerstore().AddAddrs(node1.Host.ID(), node1.Host.Addrs(), peerstore.PermanentAddrTTL)
	time.Sleep(time.Second * 3)

	//n2Info := peer.AddrInfo{
	//	ID:    node2.Host.ID(),
	//	Addrs: node2.Host.Addrs(),
	//}
	//
	//if err := node1.Host.Connect(context.TODO(), n2Info); err != nil {
	//	log.Println("error connect:", err)
	//}

	log.Println("node1 peers:", node1.Host.Peerstore().Peers())
	log.Println("node2 peers:", node2.Host.Peerstore().Peers())

	time.Sleep(10 * time.Second)
}

const (
	replyTimeout    = time.Second * 300
	ctxAlice        = "alice"
	ctxBob          = "bob"
	ctxProxy        = "proxy"
	ctxAlicePeerID  = "alice_peer_id"
	ctxBobPeerID    = "bob_peer_id"
	ctxProxyPeerID  = "proxy_peer_id"
	ctxAPoints      = "a_points"
	ctxAPrimePoints = "a_prime_points"
	ctxLPrime       = "l_prime_point"
	ctxFiles        = "files"
	ctxFilesCount   = "files_count"
)

var (
	tcSystemInit          = NewTimeCounter("sys_init")
	tcSystemShare         = NewTimeCounter("sys_share")
	tcAliceEncrypt        = NewTimeCounter("alice_encrypt")
	tcAliceCalculateBs    = NewTimeCounter("alice_calculate_bs")
	tcAliceGenerateReKeys = NewTimeCounter("alice_generate_re_keys")
	tcBobSendChoice       = NewTimeCounter("bob_send_choice")
	tcBobDecrypt          = NewTimeCounter("bob_decrypt")
	tcProxyReEncrypt      = NewTimeCounter("proxy_re_encrypt")
)

var allTimeCounters = []*TimeCounter{
	tcSystemInit,
	tcSystemShare,
	tcAliceEncrypt,
	tcAliceCalculateBs,
	tcAliceGenerateReKeys,
	tcBobSendChoice,
	tcBobDecrypt,
	tcProxyReEncrypt,
}

func ClearAllTimeCounters() {
	for _, tc := range allTimeCounters {
		tc.Clear()
	}
}

func PrintAllTimeCounters() {
	for _, tc := range allTimeCounters {
		fmt.Printf("item: %s, count: %d, avg: %d ms\n", tc.Name(), tc.Count(), tc.Avg().Milliseconds())
	}
}

type TimeCounter struct {
	name      string
	durations []time.Duration
}

func NewTimeCounter(name string) *TimeCounter {
	return &TimeCounter{name: name}
}

func (tc *TimeCounter) Name() string {
	return tc.name
}

func (tc *TimeCounter) Add(duration time.Duration) {
	tc.durations = append(tc.durations, duration)
}

func (tc *TimeCounter) Avg() time.Duration {
	var sum time.Duration
	for _, duration := range tc.durations {
		sum += duration
	}
	return sum / time.Duration(len(tc.durations))
}

func (tc *TimeCounter) Count() int {
	return len(tc.durations)
}

func (tc *TimeCounter) Durations() []time.Duration {
	return tc.durations
}

func (tc *TimeCounter) Clear() {
	tc.durations = nil
}

type Alice struct {
	Ctx *ContextWithValue
	*node.BaseNode
	*protocol.OtServer
	*protocol.PreClient
	*protocol.StorageClient
	PrivateKey *big.Int
	PublicKey  curve.Point
}

func NewAlice(node *node.BaseNode, ctx *ContextWithValue) *Alice {
	alice := &Alice{
		Ctx:           ctx,
		BaseNode:      node,
		PreClient:     protocol.NewPreClient(node),
		StorageClient: protocol.NewStorageClient(node),
	}
	alice.OtServer = protocol.NewOtServer(alice)
	var err error
	alice.PrivateKey, alice.PublicKey, err = curve.NewRandomPoint(curve.TypeG1, rand.Reader)
	if err != nil {
		panic(err)
	}
	return alice
}

func (alice *Alice) OnChoiceRequest(choice *protocol.OtChoice) protocol.Error {
	start := time.Now()

	kps, LPrime, err := ot.CalculateKeyPoints(choice.Y, choice.L, alice.PublicKey, alice.Ctx.Value(ctxFilesCount).(int64))
	if err != nil {
		return protocol.UnknownError(err.Error())
	}

	// generate bob temp keys
	keys := deriveKeysFromPoints(kps)

	tcAliceCalculateBs.Add(time.Since(start))

	start = time.Now()

	// generate re-keys
	reKeys := make([]curve.Point, len(keys))
	for i := range keys {
		reKeys[i] = pre.GenerateReKey(alice.PrivateKey, keys[i])
	}

	tcAliceGenerateReKeys.Add(time.Since(start))

	preArgs := &protocol.PreArgs{
		Cid:    choice.Cid,
		LPrime: LPrime,
		ReKeys: reKeys,
		TxID:   choice.Cid,
	}
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(replyTimeout, cancel)
	if err := alice.SendReEncrypt(ctx, alice.Ctx.Value(ctxProxyPeerID).(peer.ID), preArgs); !err.IsNil() {
		return protocol.UnknownError(fmt.Sprintf("alice send pre args error: %v", err))
	}

	return protocol.NilError()
}

type Bob struct {
	Ctx *ContextWithValue
	*node.BaseNode
	*protocol.OtClient
	*protocol.StorageClient
	PrivateKey *big.Int
	PublicKey  curve.Point
}

func NewBob(node *node.BaseNode, ctx *ContextWithValue) *Bob {
	bob := &Bob{
		Ctx:           ctx,
		BaseNode:      node,
		OtClient:      protocol.NewOtClient(node),
		StorageClient: protocol.NewStorageClient(node),
	}
	var err error
	bob.PrivateKey, bob.PublicKey, err = curve.NewRandomPoint(curve.TypeG1, rand.Reader)
	if err != nil {
		panic(err)
	}
	return bob
}

type Proxy struct {
	Ctx *ContextWithValue
	*node.BaseNode
	*protocol.PreServer
	*protocol.StorageServer
}

func NewProxy(node *node.BaseNode, ctx *ContextWithValue) *Proxy {
	proxy := &Proxy{
		Ctx:      ctx,
		BaseNode: node,
	}
	proxy.PreServer = protocol.NewPreServer(proxy)
	proxy.StorageServer = protocol.NewStorageServer(proxy)
	return proxy
}

func (proxy *Proxy) OnReEncryptRequest(args *protocol.PreArgs) protocol.Error {
	start := time.Now()

	As := proxy.Ctx.Value(ctxAPoints).([]curve.Point)
	APrimes := make([]curve.Point, len(As))
	for i := range args.ReKeys {
		APrimes[i] = pre.ReEncrypt(As[i], args.ReKeys[i])
	}

	tcProxyReEncrypt.Add(time.Since(start))

	proxy.Ctx.SetValue(ctxAPrimePoints, APrimes)
	proxy.Ctx.SetValue(ctxLPrime, args.LPrime)

	return protocol.NilError()
}

func (proxy *Proxy) OnUploadRequest(ticket *protocol.UploadTicket) (string, protocol.Error) {
	return "uploader_string", protocol.NilError()
}

func (proxy *Proxy) OnDownloadRequest(ticket *protocol.DownloadTicket) (string, protocol.Error) {
	return "downloader_string", protocol.NilError()
}

func TestNodeP2P(t *testing.T) {
	var Target int64 = 5
	for FileSize := int64(1_000); FileSize <= 1_000_000; FileSize *= 10 {
		for FileCount := int64(10); FileCount <= 100; FileCount += 10 {
			err := performNodesInteraction(FileSize, FileCount, Target)
			if err != nil {
				t.Error(err)
			}
			fmt.Println()
			fmt.Printf("FileSize: %d, FileCount: %d, Target: %d\n", FileSize, FileCount, Target)
			PrintAllTimeCounters()
			ClearAllTimeCounters()
			fmt.Println()
		}
	}
}

func performNodesInteraction(FileSize, FileCount, Target int64) error {
	alice, bob, proxy, ctxV, closer := makeThreeParty()
	defer closer()

	mh, err := multihash.FromB58String("QmQTw94j68Dgakgtfd45bG3TZG6CAfc427UVRH4mugg4q4")
	if err != nil {
		return fmt.Errorf("mh decode: %w", err)
	}

	files := mustMakeNRandomBytes(FileSize, int(FileCount))

	sysInitStart := time.Now()

	// alice encrypt files
	As, ciphers, err := encryptFiles(alice, files)
	if err != nil {
		return err
	}
	ctxV.SetValue(ctxAPoints, As)

	// alice upload encrypted files
	if err = uploadFiles(alice, proxy, ciphers, mh); err != nil {
		return err
	}
	ctxV.SetValue(ctxFiles, ciphers)
	ctxV.SetValue(ctxFilesCount, FileCount)

	tcSystemInit.Add(time.Since(sysInitStart))

	sysShareStart := time.Now()

	// bob send choice
	if err = sendChoice(alice, bob, mh, Target); err != nil {
		return err
	}

	// bob download files
	downloadTicket := &protocol.DownloadTicket{
		Owner: curve.NewPoint(curve.TypeG1).ScalarBaseMult(big.NewInt(1)),
		Cid:   cid.NewCidV0(mh),
		TxID:  cid.NewCidV0(mh),
	}
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(replyTimeout, cancel)
	if _, err := bob.SendDownloadRequest(ctx, proxy.Host.ID(), downloadTicket); err != nil {
		return fmt.Errorf("bob download ticket: %w", err)
	}

	// bob decrypt files
	bobCipher := bob.Ctx.Value(ctxFiles).([][]byte)[Target-1]
	bobFile, err := decryptFile(bob, bobCipher, Target-1)
	if err != nil {
		return fmt.Errorf("bob decrypt file: %w", err)
	}
	targetFile := files[Target-1]

	tcSystemShare.Add(time.Since(sysShareStart))

	// compare
	if !bytes.Equal(bobFile, targetFile) {
		return errors.New("bob decrypted file not equal")
	}

	return nil
}

type ContextWithValue struct {
	m *sync.Map
}

func NewContextWithValue() *ContextWithValue {
	return &ContextWithValue{m: &sync.Map{}}
}

func (ctx *ContextWithValue) Value(key interface{}) interface{} {
	value, _ := ctx.m.Load(key)
	return value
}

func (ctx *ContextWithValue) SetValue(key, value interface{}) {
	ctx.m.Store(key, value)
}

func deriveKeysFromPoints(points []curve.Point) []*big.Int {
	keys := make([]*big.Int, len(points))
	for i := range points {
		keys[i] = curve.DeriveFieldElementFromPoint(points[i])
	}
	return keys
}

func makeThreeParty() (*Alice, *Bob, *Proxy, *ContextWithValue, func()) {
	nodeAlice := node.NewRandomBaseNode(41001)
	//log.Println("running nodeAlice...", nodeAlice.Host.ID(), nodeAlice.Host.Addrs())

	nodeBob := node.NewRandomBaseNode(41002)
	//log.Println("running nodeBob...", nodeBob.Host.ID(), nodeBob.Host.Addrs())

	nodeProxy := node.NewRandomBaseNode(41003)
	//log.Println("running nodeProxy...", nodeProxy.Host.ID(), nodeProxy.Host.Addrs())

	nodeAlice.Host.Peerstore().AddAddrs(nodeBob.Host.ID(), nodeBob.Host.Addrs(), peerstore.PermanentAddrTTL)
	nodeAlice.Host.Peerstore().AddAddrs(nodeProxy.Host.ID(), nodeProxy.Host.Addrs(), peerstore.PermanentAddrTTL)

	nodeBob.Host.Peerstore().AddAddrs(nodeAlice.Host.ID(), nodeAlice.Host.Addrs(), peerstore.PermanentAddrTTL)
	nodeBob.Host.Peerstore().AddAddrs(nodeProxy.Host.ID(), nodeProxy.Host.Addrs(), peerstore.PermanentAddrTTL)

	nodeProxy.Host.Peerstore().AddAddrs(nodeAlice.Host.ID(), nodeProxy.Host.Addrs(), peerstore.PermanentAddrTTL)
	nodeProxy.Host.Peerstore().AddAddrs(nodeBob.Host.ID(), nodeBob.Host.Addrs(), peerstore.PermanentAddrTTL)

	ctx := NewContextWithValue()
	alice := NewAlice(nodeAlice, ctx)
	bob := NewBob(nodeBob, ctx)
	proxy := NewProxy(nodeProxy, ctx)

	ctx.SetValue(ctxAlice, alice)
	ctx.SetValue(ctxBob, bob)
	ctx.SetValue(ctxProxy, proxy)
	ctx.SetValue(ctxAlicePeerID, alice.Host.ID())
	ctx.SetValue(ctxBobPeerID, bob.Host.ID())
	ctx.SetValue(ctxProxyPeerID, proxy.Host.ID())

	closer := func() {
		alice.Host.Close()
		bob.Host.Close()
		proxy.Host.Close()
		//log.Println("[INFO] three party closed")
	}
	return alice, bob, proxy, ctx, closer
}

func mustMakeNRandomBytes(size int64, n int) [][]byte {
	result := make([][]byte, n)
	for i := range result {
		result[i] = mustMakeRandomBytes(size)
	}
	return result
}

func mustMakeRandomBytes(size int64) []byte {
	buf := make([]byte, size)
	if _, err := mrand.Read(buf); err != nil {
		panic(err)
	}
	return buf
}

func encryptFiles(alice *Alice, files [][]byte) (As []curve.Point, ciphertexts [][]byte, err error) {
	start := time.Now()
	for i := range files {
		a, ciphertext, err := encryptFile(alice, files[i])
		if err != nil {
			return nil, nil, fmt.Errorf("encryptFiles: %w", err)
		}
		As = append(As, a)
		ciphertexts = append(ciphertexts, ciphertext)
	}
	tcAliceEncrypt.Add(time.Since(start))
	return
}

func encryptFile(alice *Alice, file []byte) (A curve.Point, ciphertext []byte, err error) {
	cipherBuf := bytes.NewBuffer(nil)
	A, err = pre.Encrypt(alice.PublicKey, pre.NewEncryptClosure(bytes.NewReader(file), cipherBuf))
	if err != nil {
		return nil, nil, fmt.Errorf("alice encryptFile: %w", err)
	}
	ciphertext = cipherBuf.Bytes()
	return
}

func decryptFile(bob *Bob, cipher []byte, ordinal int64) (file []byte, err error) {
	start := time.Now()

	APrimes := bob.Ctx.Value(ctxAPrimePoints).([]curve.Point)
	LPrime := bob.Ctx.Value(ctxLPrime).(curve.Point)
	kp := ot.RevealKeyPoint(LPrime, bob.PrivateKey)
	b := curve.DeriveFieldElementFromPoint(kp)

	buf := bytes.NewBuffer(nil)
	err = pre.DecryptByReceiver(APrimes[ordinal], b, pre.NewDecryptClosure(bytes.NewReader(cipher), buf))
	if err != nil {
		return nil, err
	}

	tcBobDecrypt.Add(time.Since(start))
	return buf.Bytes(), nil
}

func uploadFiles(alice *Alice, proxy *Proxy, files [][]byte, mh multihash.Multihash) error {
	for i := range files {
		if err := uploadFile(alice, proxy, files[i], mh); err != nil {
			return fmt.Errorf("upload files %d: %w", i, err)
		}
	}
	return nil
}

func uploadFile(alice *Alice, proxy *Proxy, file []byte, mh multihash.Multihash) error {
	uploadTicket := &protocol.UploadTicket{
		Owner: alice.PublicKey,
		Set:   1,
		Cid:   cid.NewCidV0(mh),
		Size:  uint64(len(file)),
	}
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(replyTimeout, cancel)
	if _, err := alice.SendUploadRequest(ctx, proxy.Host.ID(), uploadTicket); err != nil {
		return fmt.Errorf("alice upload ticket: %w", err)
	}
	return nil
}

func sendChoice(alice *Alice, bob *Bob, mh multihash.Multihash, target int64) error {
	start := time.Now()

	yp, lp, err := ot.SealChoice(big.NewInt(target), alice.PublicKey, bob.PublicKey)
	if err != nil {
		return err
	}

	tcBobSendChoice.Add(time.Since(start))

	choice := &protocol.OtChoice{
		Cid:   cid.NewCidV0(mh),
		Owner: alice.PublicKey,
		Y:     yp,
		L:     lp,
	}
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(replyTimeout, cancel)
	if err := bob.SendChoice(ctx, alice.Host.ID(), choice); !err.IsNil() {
		fmt.Errorf("bob send choice: %w", err)
	}
	return nil
}

func TestGenerateKeyTime(t *testing.T) {
	var round = 100_000
	testGenerateKeyTime(round, true)
}

func testGenerateKeyTime(round int, print bool) time.Duration {
	// generate ts
	ts := make([]*big.Int, round)
	for i := range ts {
		ts[i] = big.NewInt(mrand.Int63())
	}
	// generate y points
	yPoints := make([]curve.Point, round)
	for i := range yPoints {
		yPoints[i] = newRandPoint(curve.TypeG1)
	}
	// generate pkAs
	pkAs := make([]curve.Point, round)
	for i := range pkAs {
		pkAs[i] = newRandPoint(curve.TypeG1)
	}
	// generate ordinals
	ordinals := make([]*big.Int, round)
	for i := range ordinals {
		ordinals[i] = big.NewInt(mrand.Int63n(100_000))
	}
	// generate keys
	start := time.Now()
	for i := 0; i < round; i++ {
		kpi := calculateKeyPoint(ts[i], yPoints[i], pkAs[i], ordinals[i])
		curve.DeriveFieldElementFromPoint(kpi)
	}
	elapsed := time.Since(start)
	if print {
		fmt.Printf("node: generate_key, round: %d, elapsed(ms): %d, avg(ms): %f\n",
			round, elapsed.Milliseconds(), float64(elapsed.Milliseconds())/float64(round))
	}
	return elapsed
}

func calculateKeyPoint(t *big.Int, yp, pkA curve.Point, ordinal *big.Int) (kpi curve.Point) {
	// kpi = t * yp - i * t * pkA
	kpi = curve.NewPoint(curve.TypeG1).ScalarMult(yp, t)
	temp := curve.NewPoint(curve.TypeG1).ScalarMult(pkA, new(big.Int).Mul(ordinal, t))
	kpi.Add(kpi, temp.Neg(temp))
	return
}

func newRandPoint(typ curve.Curve) curve.Point {
	r := big.NewInt(mrand.Int63())
	return curve.NewPoint(typ).ScalarBaseMult(r)
}

func init() {
	mrand.Seed(time.Now().UnixNano())
}
