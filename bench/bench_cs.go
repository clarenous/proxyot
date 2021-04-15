package bench

import (
	crand "crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/clarenous/proxyot/chash"
	"github.com/clarenous/proxyot/curve"
	"github.com/clarenous/proxyot/fileobj"
	"github.com/clarenous/proxyot/merkle"
)

// Cloud Storage Figure1 includes:
// Setup (key generation), Chameleon Hash, Merkle Tree, Update Block
type CSFigure1Result struct {
	ExecSetupMsTimes  []float64 `json:"exec_setup_ms_times"`
	ExecHashMsTimes   []float64 `json:"exec_hash_ms_times"`
	ExecMerkleMsTimes []float64 `json:"exec_merkle_ms_times"`
	ExecUpdateMsTimes []float64 `json:"exec_update_ms_times"`
	FileSizes         []int64   `json:"file_sizes"`
	BlockCounts       []int64   `json:"block_counts"`
	BlockSizes        []int64   `json:"block_sizes"`
}

// Cloud Storage Figure1 includes:
// Setup (key generation), Chameleon Hash, Merkle Tree, Update Block
func CSFigure1(rounds int, blockCounts, blockSizes []int64) (*CSFigure1Result, error) {
	result := &CSFigure1Result{
		ExecSetupMsTimes:  nil,
		ExecHashMsTimes:   nil,
		ExecMerkleMsTimes: nil,
		ExecUpdateMsTimes: nil,
		FileSizes:         nil,
		BlockCounts:       make([]int64, len(blockCounts)),
		BlockSizes:        make([]int64, len(blockSizes)),
	}
	copy(result.BlockCounts, blockCounts)
	copy(result.BlockSizes, blockSizes)
	for _, bCount := range blockCounts {
		for _, bSize := range blockSizes {
			fmt.Println("Running CSFigure1", bCount, bSize)
			var sumSetup, sumHash, sumMerkle, sumUpdate int64
			for round := 0; round < rounds; round++ {
				setupT, hashT, merkleT, updateT, err := runCSFigure1(bCount, bSize)
				if err != nil {
					return nil, err
				}
				sumSetup += setupT.Nanoseconds()
				sumHash += hashT.Nanoseconds()
				sumMerkle += merkleT.Nanoseconds()
				sumUpdate += updateT.Nanoseconds()
			}
			roundsF := float64(rounds)
			result.FileSizes = append(result.FileSizes, bCount*bSize)
			result.ExecSetupMsTimes = append(result.ExecSetupMsTimes, ns2ms(sumSetup)/roundsF)
			result.ExecHashMsTimes = append(result.ExecHashMsTimes, ns2ms(sumHash)/roundsF)
			result.ExecMerkleMsTimes = append(result.ExecMerkleMsTimes, ns2ms(sumMerkle)/roundsF)
			result.ExecUpdateMsTimes = append(result.ExecUpdateMsTimes, ns2ms(sumUpdate)/roundsF)
		}
	}
	return result, nil
}

func runCSFigure1(blockCount, blockSize int64) (setupTime, hashTime, merkleTime, updateTime time.Duration, err error) {
	fileSize := blockCount * blockSize
	// run prepare
	var mf *fileobj.MemFileObj
	if mf, err = fileobj.NewMemFileObj(fileSize, blockSize); err != nil {
		return
	}
	newBlock := make([]byte, blockSize)
	if _, err = rand.Read(newBlock); err != nil {
		return
	}
	targetBlockIdx := rand.Int63n(blockCount)
	var targetM *big.Int
	var targetCHash chash.ChameleonHash
	// run setup
	setupStart := time.Now()
	var Y curve.Point
	var x *big.Int
	if x, Y, err = curve.NewRandomPoint(curve.TypeG1, crand.Reader); err != nil {
		return
	}
	rs := make([]*big.Int, blockCount)
	Rs := make([]curve.Point, blockCount)
	for i := int64(0); i < blockCount; i++ {
		if rs[i], Rs[i], err = curve.NewRandomPoint(curve.TypeG2, crand.Reader); err != nil {
			return
		}
	}
	setupTime = time.Since(setupStart)
	// run chameleon hash
	hashStart := time.Now()
	var cHashes []chash.ChameleonHash
	for i := int64(0); i < blockCount; i++ {
		data, err := mf.GetBlock(i)
		if err != nil {
			return 0, 0, 0, 0, err
		}
		res := merkle.SHA256(data)
		resBi := new(big.Int).SetBytes(res[:])
		m := resBi.Mod(resBi, curve.Order)
		ch := chash.ComputeHash(Y, Rs[i], m)
		cHashes = append(cHashes, ch)
		if i == targetBlockIdx {
			targetM = m
			targetCHash = ch
		}
	}
	hashTime = time.Since(hashStart)
	// run Merkle Tree
	var leaves []*merkle.Hash
	for i := range cHashes {
		leaves = append(leaves, (*merkle.Hash)(&cHashes[i]))
	}
	merkleStart := time.Now()
	merkle.NewTree(leaves)
	merkleTime = time.Since(merkleStart)
	// run Update Block
	updateStart := time.Now()
	newRes := merkle.SHA256(newBlock)
	newResBi := new(big.Int).SetBytes(newRes[:])
	newM := newResBi.Mod(newResBi, curve.Order)
	newCHash, _, _ := chash.ComputeCollision(Y, x, rs[targetBlockIdx], targetM, newM, curve.Order)
	if !newCHash.Equals(targetCHash) {
		err = errors.New("invalid chameleon hash collision")
		return
	}
	updateTime = time.Since(updateStart)
	return
}

func getBlockCount(fSize, bSize int64) int64 {
	count, rem := fSize/bSize, fSize%bSize
	if rem > 0 {
		count += 1
	}
	return count
}
