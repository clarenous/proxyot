package bench

import (
	"encoding/hex"
	"fmt"

	"github.com/clarenous/proxyot/curve"
	msg "github.com/clarenous/proxyot/node/protocol/pb"
	"github.com/golang/protobuf/proto"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

func AliceCommunicationCostResults(sizeMin, sizeMax, sizeRatio, countMin, countMax, countStep int64) (
	results []int64, sizes []int64, counts []int64) {
	for size := sizeMin; size <= sizeMax; size *= sizeRatio {
		sizes = append(sizes, size)
	}
	for count := countMin; count <= countMax; count += countStep {
		counts = append(counts, count)
	}
	for size := sizeMin; size <= sizeMax; size *= sizeRatio {
		for count := countMin; count <= countMax; count += countStep {
			results = append(results, aliceCommunicationCost(count))
			fmt.Println(size, count, results[len(results)-1])
		}
	}
	return
}

func aliceCommunicationCost(messageNumber int64) (cost int64) {
	// receive
	recv := mockOtChoiceRequest()
	recvBytes, err := marshalProtoMsg(recv)
	if err != nil {
		panic(err)
	}
	// send
	send := mockReEncryptRequest(messageNumber)
	sendbytes, err := marshalProtoMsg(send)
	if err != nil {
		panic(err)
	}
	return int64(len(recvBytes)) + int64(len(sendbytes))
}

func mockOtChoiceRequest() *msg.OtChoiceRequest {
	return &msg.OtChoiceRequest{
		Cid:   mockedCid.String(),
		Owner: mockedG1Point.Marshal(),
		Yp:    mockedG1Point.Marshal(),
		Lp:    mockedG1Point.Marshal(),
	}
}

func mockReEncryptRequest(number int64) *msg.PreReEncryptRequest {
	reKeys := make([][]byte, number)
	for i := range reKeys {
		reKeys[i] = mockedG2Point.Marshal()
	}
	return &msg.PreReEncryptRequest{
		Cid:    mockedCid.String(),
		Lpp:    mockedG1Point.Marshal(),
		ReKeys: reKeys,
		Txid:   mockedCid.String(),
	}
}

func marshalProtoMsg(pb proto.Message) ([]byte, error) {
	return proto.Marshal(pb)
}

var (
	mockedMultiHash multihash.Multihash
	mockedCid       cid.Cid
	mockedG1Point   = randPoint(curve.TypeG1)
	mockedG2Point   = randPoint(curve.TypeG2)
)

func init() {
	buf, err := hex.DecodeString("e59d4ffa01131dd5746a32b8b032a4cfb82f73ef09e59b2771990dd11a12e206")
	if err != nil {
		panic(err)
	}
	data, err := multihash.Encode(buf, multihash.SHA2_256)
	if err != nil {
		panic(err)
	}
	mockedMultiHash, err = multihash.Cast(data)
	if err != nil {
		panic(err)
	}
	mockedCid = cid.NewCidV0(mockedMultiHash)
}
