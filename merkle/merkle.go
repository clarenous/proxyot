package merkle

import (
	"crypto/sha256"
	"encoding/hex"
)

type Hash [sha256.Size]byte

func (h *Hash) Array() [sha256.Size]byte {
	return *h
}

func (h *Hash) Bytes() []byte {
	bs := h.Array()
	return bs[:]
}

func (h *Hash) String() string {
	return hex.EncodeToString((*h)[:])
}

func (h Hash) Ptr() *Hash {
	return &h
}

func SHA256(data []byte) Hash {
	return sha256.Sum256(data)
}

type Tree struct {
	Root   *Hash
	Leaves []*Hash
}

func NewTree(leaves []*Hash) *Tree {
	if len(leaves) == 0 {
		return nil
	}
	layer := make([]*Hash, len(leaves))
	copy(layer, leaves)
	for len(layer) > 1 {
		layer = computeLayer(layer)
	}
	return &Tree{
		Root:   SHA256(layer[0].Bytes()).Ptr(),
		Leaves: leaves,
	}
}

func computeLayer(nodes []*Hash) (results []*Hash) {
	count, rem := len(nodes)/2, len(nodes)%2
	for i := 0; i < count; i++ {
		results = append(results, computePair(nodes[i*2], nodes[i*2+1]))
	}
	if rem == 1 {
		results = append(results, SHA256(nodes[count*2].Bytes()).Ptr())
	}
	return results
}

func computePair(left, right *Hash) *Hash {
	data := make([]byte, sha256.Size*2)
	copy(data, left.Bytes())
	copy(data[sha256.Size:], right.Bytes())
	return SHA256(data).Ptr()
}
