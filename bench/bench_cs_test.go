package bench_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/clarenous/proxyot/bench"
)

const (
	KiB = 1024
	MiB = 1024 * KiB
)

func TestCSFigure1(t *testing.T) {
	var rounds = 100
	blockCounts := []int64{16, 64, 256}
	blockSizes := []int64{64 * KiB, 256 * KiB, MiB}
	result, err := bench.CSFigure1(rounds, blockCounts, blockSizes)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
}

func TestCSFigure2(t *testing.T) {
	var rounds = 100
	var updateCount int64 = 10
	blockSizes := []int64{64 * KiB, 256 * KiB, MiB}
	result, err := bench.CSFigure2(rounds, updateCount, blockSizes)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
}
