package merkle_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/clarenous/proxyot/merkle"
)

func TestNewTree(t *testing.T) {
	var leavesCount = []int{1, 2, 3, 4, 8, 9, 15, 16}
	for _, count := range leavesCount {
		leaves := generateLeaves(count)
		tree := merkle.NewTree(leaves)
		t.Log(count, tree.Root.String())
	}
}

func generateLeaves(n int) (leaves []*merkle.Hash) {
	if n <= 0 {
		return
	}
	for i := 0; i < n; i++ {
		leaves = append(leaves, merkle.SHA256(mustRead64Bytes()).Ptr())
	}
	return
}

func mustRead64Bytes() (bs []byte) {
	bs = make([]byte, 64)
	if _, err := rand.Read(bs); err != nil {
		panic(err)
	}
	return
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
