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
	otChoiceRequest  = "/proxyot/ot/choicereq/0.0.1"  // bob send sealed choice to alice
	otChoiceResponse = "/proxyot/ot/choiceresp/0.0.1" // alice reply
)

type OtChoice struct {
	Cid   cid.Cid
	Owner curve.Point
	Y     curve.Point
	L     curve.Point
}

type OtServerNode interface {
	BaseNode
	OnChoiceRequest(choice *OtChoice) Error
}

func NewOtServer(node OtServerNode) *OtServer {
	server := &OtServer{node: node}
	node.SetStreamHandler(otChoiceRequest, server.onChoiceRequest)
	return server
}

// OtServer type
type OtServer struct {
	node OtServerNode
}

func (p *OtServer) onChoiceRequest(s network.Stream) {
	// read msg
	req := &msg.OtChoiceRequest{}
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

	owner := curve.NewPoint(curve.TypeG1)
	if _, err := owner.Unmarshal(req.Owner); err != nil {
		// TODO: bad owner
		return
	}

	yp := curve.NewPoint(curve.TypeG1)
	if _, err = yp.Unmarshal(req.Yp); err != nil {
		// TODO: bad Y point
		return
	}

	lp := curve.NewPoint(curve.TypeG1)
	if _, err = lp.Unmarshal(req.Lp); err != nil {
		// TODO: bad Y point
		return
	}

	choice := &OtChoice{
		Cid:   id,
		Owner: owner,
		Y:     yp,
		L:     lp,
	}
	pErr := p.node.OnChoiceRequest(choice)

	resp := &msg.OtChoiceResponse{
		ErrorCode: pErr.Code(),
		ErrorMsg:  pErr.Msg(),
	}
	sendProtoMsg(p.node, s.Conn().RemotePeer(), otChoiceResponse, resp)
	// TODO: response failed ?
}

type OtClient struct {
	node BaseNode
	pool *MessagePool
}

func NewOtClient(node BaseNode) *OtClient {
	client := &OtClient{
		node: node,
		pool: NewMessagePool(),
	}
	client.node.SetStreamHandler(otChoiceResponse, client.onChoiceResponse)
	return client
}

// TODO: check result
func (p *OtClient) onChoiceResponse(s network.Stream) {
	// read msg
	resp := &msg.OtChoiceResponse{}
	if err := readMsgFromStream(resp, s); err != nil {
		s.Reset()
		log.Println(err)
		return
	}
	s.Close()
	err := p.pool.Push(otChoiceResponse, resp)
	fmt.Println("Ot Choice Response:", resp.ErrorCode, resp.ErrorMsg, err)
}

func (p *OtClient) SendChoice(ctx context.Context, peerID peer.ID, choice *OtChoice) Error {
	req := &msg.OtChoiceRequest{
		Cid:   choice.Cid.String(),
		Owner: choice.Owner.Marshal(),
		Yp:    choice.Y.Marshal(),
		Lp:    choice.L.Marshal(),
	}

	sendProtoMsg(p.node, peerID, otChoiceRequest, req)
	v, err := p.pool.Wait(ctx, otChoiceResponse)
	if err != nil {
		return UnknownError(err.Error())
	}
	resp := v.(*msg.OtChoiceResponse)
	return NewError(resp.ErrorCode, resp.ErrorMsg)
}
