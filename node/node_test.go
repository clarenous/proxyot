package node_test

import (
	"context"
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/clarenous/proxyot/curve"
	"github.com/clarenous/proxyot/node"
	"github.com/clarenous/proxyot/node/protocol"
	"github.com/ipfs/go-cid"
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

type Alice struct {
	*node.BaseNode
	*protocol.OtServer
	*protocol.PreClient
	*protocol.StorageClient
}

func NewAlice(node *node.BaseNode) *Alice {
	return &Alice{
		BaseNode:      node,
		OtServer:      protocol.NewOtServer(node),
		PreClient:     protocol.NewPreClient(node),
		StorageClient: protocol.NewStorageClient(node),
	}
}

type Bob struct {
	*node.BaseNode
	*protocol.OtClient
	*protocol.StorageClient
}

func NewBob(node *node.BaseNode) *Bob {
	return &Bob{
		BaseNode:      node,
		OtClient:      protocol.NewOtClient(node),
		StorageClient: protocol.NewStorageClient(node),
	}
}

type Proxy struct {
	*node.BaseNode
	*protocol.PreServer
	*protocol.StorageServer
}

func NewProxy(node *node.BaseNode) *Proxy {
	return &Proxy{
		BaseNode:      node,
		PreServer:     protocol.NewPreServer(node),
		StorageServer: protocol.NewStorageServer(node),
	}
}

func TestNodeP2P(t *testing.T) {
	nodeAlice := node.NewRandomBaseNode(41001)
	log.Println("running nodeAlice...", nodeAlice.Host.ID(), nodeAlice.Host.Addrs())
	defer nodeAlice.Host.Close()

	nodeBob := node.NewRandomBaseNode(41002)
	log.Println("running nodeBob...", nodeBob.Host.ID(), nodeBob.Host.Addrs())
	defer nodeBob.Host.Close()

	nodeProxy := node.NewRandomBaseNode(41003)
	log.Println("running nodeProxy...", nodeProxy.Host.ID(), nodeProxy.Host.Addrs())
	defer nodeProxy.Host.Close()

	nodeAlice.Host.Peerstore().AddAddrs(nodeBob.Host.ID(), nodeBob.Host.Addrs(), peerstore.PermanentAddrTTL)
	nodeAlice.Host.Peerstore().AddAddrs(nodeProxy.Host.ID(), nodeProxy.Host.Addrs(), peerstore.PermanentAddrTTL)

	nodeBob.Host.Peerstore().AddAddrs(nodeAlice.Host.ID(), nodeAlice.Host.Addrs(), peerstore.PermanentAddrTTL)
	nodeBob.Host.Peerstore().AddAddrs(nodeProxy.Host.ID(), nodeProxy.Host.Addrs(), peerstore.PermanentAddrTTL)

	nodeProxy.Host.Peerstore().AddAddrs(nodeAlice.Host.ID(), nodeProxy.Host.Addrs(), peerstore.PermanentAddrTTL)
	nodeProxy.Host.Peerstore().AddAddrs(nodeBob.Host.ID(), nodeBob.Host.Addrs(), peerstore.PermanentAddrTTL)

	alice := NewAlice(nodeAlice)
	bob := NewBob(nodeBob)
	_ = NewProxy(nodeProxy)

	mh, err := multihash.FromB58String("QmQTw94j68Dgakgtfd45bG3TZG6CAfc427UVRH4mugg4q4")
	if err != nil {
		t.Fatal("mh decode", err)
	}

	var ctx context.Context
	var cancel context.CancelFunc

	uploadTicket := &protocol.UploadTicket{
		Owner: curve.NewPoint(curve.TypeG1).ScalarBaseMult(big.NewInt(1)),
		Set:   1,
		Cid:   cid.NewCidV0(mh),
		Size:  1024,
	}
	ctx, cancel = context.WithCancel(context.Background())
	time.AfterFunc(time.Second*5, cancel)
	if _, err := alice.SendUploadRequest(ctx, nodeProxy.Host.ID(), uploadTicket); err != nil {
		t.Error("alice upload ticket error:", err)
	}

	choice := &protocol.OtChoice{
		Cid:   cid.NewCidV0(mh),
		Owner: curve.NewPoint(curve.TypeG1).ScalarBaseMult(big.NewInt(1)),
		Y:     curve.NewPoint(curve.TypeG1).ScalarBaseMult(big.NewInt(2)),
		L:     curve.NewPoint(curve.TypeG1).ScalarBaseMult(big.NewInt(3)),
	}
	ctx, cancel = context.WithCancel(context.Background())
	time.AfterFunc(time.Second*5, cancel)
	if err := bob.SendChoice(ctx, nodeAlice.Host.ID(), choice); !err.IsNil() {
		t.Error("bob send choice error:", err)
	}

	preArgs := &protocol.PreArgs{
		Cid:   cid.NewCidV0(mh),
		A:     curve.NewPoint(curve.TypeG1).ScalarBaseMult(big.NewInt(4)),
		ReKey: curve.NewPoint(curve.TypeG2).ScalarBaseMult(big.NewInt(5)),
		TxID:  cid.NewCidV0(mh),
	}
	ctx, cancel = context.WithCancel(context.Background())
	time.AfterFunc(time.Second*5, cancel)
	if err := alice.SendReEncrypt(ctx, nodeProxy.Host.ID(), preArgs); !err.IsNil() {
		t.Error("alice send pre args error:", err)
	}

	downloadTicket := &protocol.DownloadTicket{
		Owner: curve.NewPoint(curve.TypeG1).ScalarBaseMult(big.NewInt(1)),
		Cid:   cid.NewCidV0(mh),
		TxID:  cid.NewCidV0(mh),
	}
	ctx, cancel = context.WithCancel(context.Background())
	time.AfterFunc(time.Second*5, cancel)
	if _, err := bob.SendDownloadRequest(ctx, nodeProxy.Host.ID(), downloadTicket); err != nil {
		t.Error("bob download ticket error:", err)
	}
}
