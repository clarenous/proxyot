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
	storUploadRequest    = "/proxyot/stor/uploadreq/0.0.1"    // alice asks for file upload handler
	storUploadResponse   = "/proxyot/stor/uploadresp/0.0.1"   // proxy reply
	storDownloadRequest  = "/proxyot/stor/downloadreq/0.0.1"  // alice/bob asks for file download handler
	storDownloadResponse = "/proxyot/stor/downloadresp/0.0.1" // proxy reply
)

type UploadTicket struct {
	Owner curve.Point
	Set   uint32
	Cid   cid.Cid
	Size  uint64
}

type DownloadTicket struct {
	Owner curve.Point
	Cid   cid.Cid
	TxID  cid.Cid
}

type StorageServerNode interface {
	BaseNode
	OnUploadRequest(ticket *UploadTicket) (string, Error)
	OnDownloadRequest(ticket *DownloadTicket) (string, Error)
}

// StorageServer type
type StorageServer struct {
	node StorageServerNode
}

func NewStorageServer(node StorageServerNode) *StorageServer {
	server := &StorageServer{node: node}
	server.node.SetStreamHandler(storUploadRequest, server.onUploadRequest)
	server.node.SetStreamHandler(storDownloadRequest, server.onDownloadRequest)
	return server
}

func (p *StorageServer) onUploadRequest(s network.Stream) {
	// read msg
	req := &msg.StorUploadRequest{}
	if err := readMsgFromStream(req, s); err != nil {
		s.Reset()
		log.Println(err)
		return
	}
	s.Close()

	// check msg
	owner := curve.NewPoint(curve.TypeG1)
	if _, err := owner.Unmarshal(req.Owner); err != nil {
		// TODO: bad owner
		return
	}

	id, err := cid.Decode(req.Cid)
	if err != nil {
		// TODO: wrong cid
		return
	}

	ticket := &UploadTicket{
		Owner: owner,
		Set:   req.Set,
		Cid:   id,
		Size:  req.FileSize,
	}

	uploader, pErr := p.node.OnUploadRequest(ticket)

	resp := &msg.StorUploadResponse{
		Uploader:  uploader,
		ErrorCode: pErr.Code(),
		ErrorMsg:  pErr.Msg(),
	}
	sendProtoMsg(p.node, s.Conn().RemotePeer(), storUploadResponse, resp)
	// TODO: response failed ?
}

func (p *StorageServer) onDownloadRequest(s network.Stream) {
	// read msg
	req := &msg.StorDownloadRequest{}
	if err := readMsgFromStream(req, s); err != nil {
		s.Reset()
		log.Println(err)
		return
	}
	s.Close()

	// check msg
	owner := curve.NewPoint(curve.TypeG1)
	if _, err := owner.Unmarshal(req.Owner); err != nil {
		// TODO: bad owner
		return
	}

	id, err := cid.Decode(req.Cid)
	if err != nil {
		// TODO: wrong cid
		return
	}

	txid, err := cid.Decode(req.Txid)
	if err != nil {
		// TODO: wrong txid
		return
	}

	ticket := &DownloadTicket{
		Owner: owner,
		Cid:   id,
		TxID:  txid,
	}

	downloader, pErr := p.node.OnDownloadRequest(ticket)

	resp := &msg.StorDownloadResponse{
		Downloader: downloader,
		ErrorCode:  pErr.Code(),
		ErrorMsg:   pErr.Msg(),
	}
	sendProtoMsg(p.node, s.Conn().RemotePeer(), storDownloadResponse, resp)
	// TODO: response failed ?
}

// StorageClient type
type StorageClient struct {
	node BaseNode
	pool *MessagePool
}

func NewStorageClient(node BaseNode) *StorageClient {
	client := &StorageClient{
		node: node,
		pool: NewMessagePool(),
	}
	client.node.SetStreamHandler(storUploadResponse, client.onUploadResponse)
	client.node.SetStreamHandler(storDownloadResponse, client.onDownloadResponse)
	return client
}

func (p *StorageClient) onUploadResponse(s network.Stream) {
	// read msg
	resp := &msg.StorUploadResponse{}
	if err := readMsgFromStream(resp, s); err != nil {
		s.Reset()
		log.Println(err)
		return
	}
	s.Close()
	if err := p.pool.Push(storUploadResponse, resp); err != nil {
		fmt.Println("Stor Upload Response:", resp.ErrorCode, resp.ErrorMsg, err)
	}
}

func (p *StorageClient) onDownloadResponse(s network.Stream) {
	// read msg
	resp := &msg.StorDownloadResponse{}
	if err := readMsgFromStream(resp, s); err != nil {
		s.Reset()
		log.Println(err)
		return
	}
	s.Close()
	if err := p.pool.Push(storDownloadResponse, resp); err != nil {
		fmt.Println("Stor Download Response:", resp.ErrorCode, resp.ErrorMsg, err)
	}
}

func (p *StorageClient) SendUploadRequest(ctx context.Context, peerID peer.ID, ticket *UploadTicket) (string, error) {
	req := &msg.StorUploadRequest{
		Owner:    ticket.Owner.Marshal(),
		Set:      ticket.Set,
		Cid:      ticket.Cid.String(),
		FileSize: ticket.Size,
	}
	sendProtoMsg(p.node, peerID, storUploadRequest, req)
	v, err := p.pool.Wait(ctx, storUploadResponse)
	if err != nil {
		return "", err
	}
	resp := v.(*msg.StorUploadResponse)
	if resp.ErrorCode != 0 {
		err = NewError(resp.ErrorCode, resp.ErrorMsg)
	}
	return resp.Uploader, err
}

func (p *StorageClient) SendDownloadRequest(ctx context.Context, peerID peer.ID, ticket *DownloadTicket) (string, error) {
	req := &msg.StorDownloadRequest{
		Owner: ticket.Owner.Marshal(),
		Cid:   ticket.Cid.String(),
		Txid:  ticket.TxID.String(),
	}
	sendProtoMsg(p.node, peerID, storDownloadRequest, req)
	v, err := p.pool.Wait(ctx, storDownloadResponse)
	if err != nil {
		return "", err
	}
	resp := v.(*msg.StorDownloadResponse)
	if resp.ErrorCode != 0 {
		err = NewError(resp.ErrorCode, resp.ErrorMsg)
	}
	return resp.Downloader, err
}
