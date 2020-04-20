package node

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	p2pprotocol "github.com/libp2p/go-libp2p-core/protocol"
	ma "github.com/multiformats/go-multiaddr"
)

type BaseNode struct {
	ListenAddress ma.Multiaddr
	P2PPrivateKey crypto.PrivKey
	Host          host.Host
}

func NewRandomBaseNode(port uint16) *BaseNode {
	listen, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
	priv, _, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	nodeHost, _ := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(listen),
		libp2p.Identity(priv),
	)
	return &BaseNode{
		ListenAddress: listen,
		P2PPrivateKey: priv,
		Host:          nodeHost,
	}
}

func (node *BaseNode) NewStream(ctx context.Context, p peer.ID, pids ...p2pprotocol.ID) (network.Stream, error) {
	return node.Host.NewStream(ctx, p, pids...)
}

func (node *BaseNode) SetStreamHandler(pid p2pprotocol.ID, handler network.StreamHandler) {
	node.Host.SetStreamHandler(pid, handler)
}
