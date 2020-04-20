package protocol

import (
	"context"
	"fmt"
	"log"

	"github.com/clarenous/proxyot/curve"
	msg "github.com/clarenous/proxyot/node/protocol/pb"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	preReEncryptRequest  = "/proxyot/pre/reencryptreq/0.0.1"  // alice send re-key to proxy
	preReEncryptResponse = "/proxyot/pre/reencryptresp/0.0.1" // proxy reply
)

type PreArgs struct {
	Cid    cid.Cid
	LPrime curve.Point
	ReKeys []curve.Point
	TxID   cid.Cid
}

type PreServerNode interface {
	BaseNode
	OnReEncryptRequest(args *PreArgs) Error
}

func NewPreServer(node PreServerNode) *PreServer {
	server := &PreServer{node: node}
	server.node.SetStreamHandler(preReEncryptRequest, server.onReEncryptRequest)
	return server
}

// PreServer type
type PreServer struct {
	node PreServerNode
}

func (p *PreServer) onReEncryptRequest(s network.Stream) {
	// read msg
	req := &msg.PreReEncryptRequest{}
	if err := readMsgFromStream(req, s); err != nil {
		s.Reset()
		log.Println(err)
		return
	}
	s.Close()

	// check msg
	id, err := cid.Decode(req.Cid)
	if err != nil {
		// TODO: wrong cid
		return
	}

	lpp := curve.NewPoint(curve.TypeG1)
	if _, err := lpp.Unmarshal(req.Lpp); err != nil {
		// TODO: bad lpp
		return
	}

	reKeys := make([]curve.Point, len(req.ReKeys))
	for i := range req.ReKeys {
		reKey := curve.NewPoint(curve.TypeG2)
		if _, err = reKey.Unmarshal(req.ReKeys[i]); err != nil {
			// TODO: bad reKey
			return
		}
		reKeys[i] = reKey
	}

	txid, err := cid.Decode(req.Txid)
	if err != nil {
		// TODO: wrong txid
		return
	}

	args := &PreArgs{
		Cid:    id,
		LPrime: lpp,
		ReKeys: reKeys,
		TxID:   txid,
	}
	pErr := p.node.OnReEncryptRequest(args)

	resp := &msg.PreReEncryptResponse{
		ErrorCode: pErr.Code(),
		ErrorMsg:  pErr.Msg(),
	}
	sendProtoMsg(p.node, s.Conn().RemotePeer(), preReEncryptResponse, resp)
	// TODO: response failed ?
}

// PreClient type
type PreClient struct {
	node BaseNode
	pool *MessagePool
}

func NewPreClient(node BaseNode) *PreClient {
	client := &PreClient{
		node: node,
		pool: NewMessagePool(),
	}
	client.node.SetStreamHandler(preReEncryptResponse, client.onReEncryptResponse)
	return client
}

// TODO: check result
func (p *PreClient) onReEncryptResponse(s network.Stream) {
	// read msg
	resp := &msg.PreReEncryptResponse{}
	if err := readMsgFromStream(resp, s); err != nil {
		s.Reset()
		log.Println(err)
		return
	}
	s.Close()
	err := p.pool.Push(preReEncryptResponse, resp)
	fmt.Println("Pre ReEncrypt Response:", resp.ErrorCode, resp.ErrorMsg, err)
}

func (p *PreClient) SendReEncrypt(ctx context.Context, peerID peer.ID, args *PreArgs) Error {
	reKeys := make([][]byte, len(args.ReKeys))
	for i := range args.ReKeys {
		reKeys[i] = args.ReKeys[i].Marshal()
	}
	req := &msg.PreReEncryptRequest{
		Cid:    args.Cid.String(),
		Lpp:    args.LPrime.Marshal(),
		ReKeys: reKeys,
		Txid:   args.TxID.String(),
	}
	sendProtoMsg(p.node, peerID, preReEncryptRequest, req)
	v, err := p.pool.Wait(ctx, preReEncryptResponse)
	if err != nil {
		return UnknownError(err.Error())
	}
	resp := v.(*msg.PreReEncryptResponse)
	return NewError(resp.ErrorCode, resp.ErrorMsg)
}
