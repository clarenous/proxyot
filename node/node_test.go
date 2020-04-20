package node_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
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
	replyTimeout    = time.Second * 30
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
	log.Println("[INFO] OnChoiceRequest", choice)

	kps, LPrime, err := ot.CalculateKeyPoints(choice.Y, choice.L, alice.PublicKey, alice.Ctx.Value(ctxFilesCount).(int64))
	if err != nil {
		return protocol.UnknownError(err.Error())
	}

	// generate bob temp keys
	keys := deriveKeysFromPoints(kps)

	// generate re-keys
	reKeys := make([]curve.Point, len(keys))
	for i := range keys {
		reKeys[i] = pre.GenerateReKey(alice.PrivateKey, keys[i])
	}

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
	log.Println("[INFO] OnReEncryptRequest", args)

	As := proxy.Ctx.Value(ctxAPoints).([]curve.Point)
	APrimes := make([]curve.Point, len(As))
	for i := range args.ReKeys {
		APrimes[i] = pre.ReEncrypt(As[i], args.ReKeys[i])
	}

	proxy.Ctx.SetValue(ctxAPrimePoints, APrimes)
	proxy.Ctx.SetValue(ctxLPrime, args.LPrime)

	return protocol.NilError()
}

func (proxy *Proxy) OnUploadRequest(ticket *protocol.UploadTicket) (string, protocol.Error) {
	log.Println("[INFO] OnUploadRequest", ticket)
	return "uploader_string", protocol.NilError()
}

func (proxy *Proxy) OnDownloadRequest(ticket *protocol.DownloadTicket) (string, protocol.Error) {
	log.Println("[INFO] OnDownloadRequest", ticket)
	return "downloader_string", protocol.NilError()
}

func TestNodeP2P(t *testing.T) {
	var FileSize, FileCount, Target int64 = 10240, 10, 3

	alice, bob, proxy, ctxV, closer := makeThreeParty()
	defer closer()

	mh, err := multihash.FromB58String("QmQTw94j68Dgakgtfd45bG3TZG6CAfc427UVRH4mugg4q4")
	if err != nil {
		t.Fatal("mh decode", err)
	}

	files := mustMakeNRandomBytes(FileSize, int(FileCount))

	// alice encrypt files
	As, ciphers, err := encryptFiles(alice, files)
	if err != nil {
		t.Fatal(err)
	}
	ctxV.SetValue(ctxAPoints, As)

	// alice upload encrypted files
	if err = uploadFiles(alice, proxy, ciphers, mh); err != nil {
		t.Fatal(err)
	}
	ctxV.SetValue(ctxFiles, ciphers)
	ctxV.SetValue(ctxFilesCount, FileCount)

	// bob send choice
	if err = sendChoice(alice, bob, mh, Target); err != nil {
		t.Fatal(err)
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
		t.Error("bob download ticket error:", err)
	}

	// bob decrypt files
	bobCipher := bob.Ctx.Value(ctxFiles).([][]byte)[Target-1]
	bobFile, err := decryptFile(bob, bobCipher, Target-1)
	if err != nil {
		t.Fatal("bob decrypt file error:", err)
	}
	targetFile := files[Target-1]

	// compare
	if !bytes.Equal(bobFile, targetFile) {
		t.Error("file not equal")
	}
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
		keys[i] = deriveKeyFromPoint(points[i])
	}
	return keys
}

func deriveKeyFromPoint(point curve.Point) *big.Int {
	h := sha256.Sum256(point.Marshal())
	bi := new(big.Int).SetBytes(h[:])
	bi.Mod(bi, curve.Order)
	return bi
}

func makeThreeParty() (*Alice, *Bob, *Proxy, *ContextWithValue, func()) {
	nodeAlice := node.NewRandomBaseNode(41001)
	log.Println("running nodeAlice...", nodeAlice.Host.ID(), nodeAlice.Host.Addrs())

	nodeBob := node.NewRandomBaseNode(41002)
	log.Println("running nodeBob...", nodeBob.Host.ID(), nodeBob.Host.Addrs())

	nodeProxy := node.NewRandomBaseNode(41003)
	log.Println("running nodeProxy...", nodeProxy.Host.ID(), nodeProxy.Host.Addrs())

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
		log.Println("[INFO] three party closed")
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
	for i := range files {
		a, ciphertext, err := encryptFile(alice, files[i])
		if err != nil {
			return nil, nil, fmt.Errorf("encryptFiles: %w", err)
		}
		As = append(As, a)
		ciphertexts = append(ciphertexts, ciphertext)
	}
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
	APrimes := bob.Ctx.Value(ctxAPrimePoints).([]curve.Point)
	LPrime := bob.Ctx.Value(ctxLPrime).(curve.Point)
	kp := ot.RevealKeyPoint(LPrime, bob.PrivateKey)
	b := deriveKeyFromPoint(kp)

	buf := bytes.NewBuffer(nil)
	err = pre.DecryptByReceiver(APrimes[ordinal], b, pre.NewDecryptClosure(bytes.NewReader(cipher), buf))
	if err != nil {
		return nil, err
	}
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
	yp, lp, err := ot.SealChoice(big.NewInt(target), alice.PublicKey, bob.PublicKey)
	if err != nil {
		return err
	}

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
