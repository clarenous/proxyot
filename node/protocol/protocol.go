package protocol

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	ggio "github.com/gogo/protobuf/io"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/helpers"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
)

type BaseNode interface {
	NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error)
	SetStreamHandler(pid protocol.ID, handler network.StreamHandler)
}

type Error interface {
	Code() uint32
	Msg() string
	Error() string
	IsNil() bool
}

func NewError(code uint32, msg string) Error {
	return &pError{
		code: code,
		msg:  msg,
	}
}

func UnknownError(msg string) Error {
	return &pError{
		code: 1,
		msg:  msg,
	}
}

func NilError() Error {
	return &pError{}
}

type pError struct {
	code uint32
	msg  string
}

func (err *pError) Code() uint32 {
	return err.code
}

func (err *pError) Msg() string {
	return err.msg
}

func (err *pError) IsNil() bool {
	return err.code == 0
}

func (err *pError) Error() string {
	if err.code == 0 {
		return "<nil>"
	}
	return fmt.Sprintf("%d: %s", err.code, err.msg)
}

type MessagePool struct {
	l sync.RWMutex
	m map[string]chan interface{}
}

func NewMessagePool() *MessagePool {
	return &MessagePool{m: make(map[string]chan interface{})}
}

func (mp *MessagePool) Wait(ctx context.Context, protocol string) (interface{}, error) {
	mp.l.Lock()
	if _, ok := mp.m[protocol]; ok {
		mp.l.Unlock()
		return nil, errors.New("message pool msg should not exists")
	}
	ch := make(chan interface{}, 1)
	mp.m[protocol] = ch
	mp.l.Unlock()

	defer func() {
		mp.l.Lock()
		delete(mp.m, protocol)
		mp.l.Unlock()
	}()

	select {
	case <-ctx.Done():
		return nil, errors.New("message pool timeout")

	case msg := <-ch:
		return msg, nil
	}
}

func (mp *MessagePool) Push(protocol string, msg interface{}) error {
	mp.l.RLock()
	defer mp.l.RUnlock()

	ch, ok := mp.m[protocol]
	if !ok {
		return errors.New("message pool no request")
	}
	ch <- msg
	close(ch)
	return nil
}

func readMsgFromStream(msg proto.Message, s network.Stream) error {
	// read msg data from stream
	buf, err := ioutil.ReadAll(s)
	if err != nil {
		return err
	}
	// unmarshal req data
	return proto.Unmarshal(buf, msg)
}

func sendProtoMsg(bn BaseNode, id peer.ID, p protocol.ID, msg proto.Message) bool {
	s, err := bn.NewStream(context.Background(), id, p)
	if err != nil {
		log.Println("new stream:", err, p)
		return false
	}
	writer := ggio.NewFullWriter(s)
	err = writer.WriteMsg(msg)
	if err != nil {
		log.Println("write msg:", err, p)
		s.Reset()
		return false
	}
	// FullClose closes the stream and waits for the other side to close their half.
	err = helpers.FullClose(s)
	if err != nil {
		log.Println("full close:", err, p)
		s.Reset()
		return false
	}
	return true
}
