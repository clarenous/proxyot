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
	fileSizes := []int64{16 * MiB, 64 * MiB, 256 * MiB}
	blockCounts := []int64{16, 64, 256}
	result, err := bench.CSFigure1(rounds, fileSizes, blockCounts)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
}
