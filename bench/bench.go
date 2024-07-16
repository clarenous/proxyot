package bench

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/clarenous/proxyot/curve"
)

const (
	namePointMultG1   = "point_multiplication_g1"
	namePointDivG1    = "point_div_g1"
	namePointDivG2    = "point_div_g2"
	namePointDivGT    = "point_div_gt"
	namePointMultG2   = "point_multiplication_g2"
	namePointMultGT   = "point_multiplication_gt"
	namePointPair     = "point_pair"
	nameGenerateKey   = "generate_key"
	nameGenerateReKey = "generate_re_key"
	nameReEncrypt     = "re_encrypt"
	nameEncrypt       = "encrypt"
	nameDecrypt       = "decrypt"
	nameShareMessage  = "share_message"
)

type ExecResult struct {
	name  string
	bs    testing.BenchmarkResult
	extra map[string]interface{}
	str   string
}

func (r ExecResult) NsPerOp() int64 {
	return r.bs.NsPerOp()
}

func (r ExecResult) MsPerOp() float64 {
	return ns2ms(r.bs.NsPerOp())
}

func (r ExecResult) String() string {
	if r.str != "" {
		return r.str
	}
	extraKeys := make([]string, 0, len(r.extra))
	for key := range r.extra {
		extraKeys = append(extraKeys, key)
	}
	sort.Strings(extraKeys)
	extraValues := make([]string, 0, len(r.extra))
	for _, key := range extraKeys {
		extraValues = append(extraValues, fmt.Sprintf("%s: %v", key, r.extra[key]))
	}
	r.str = fmt.Sprintf("{name: %s, ms_per_op: %.3f, round: %d, %s}",
		r.name, r.MsPerOp(), r.bs.N, strings.Join(extraValues, ", "))
	return r.str
}

func ns2ms(ns int64) (ms float64) {
	return float64(ns) / 1e6
}

func ExecutionResults() []ExecResult {
	results := make([]ExecResult, 0)
	// point multiplication
	results = append(results,
		ExecResult{
			name: namePointMultG1,
			bs:   testing.Benchmark(func(b *testing.B) { benchScalarMult(b, curve.TypeG1) }),
		},
	)
	results = append(results,
		ExecResult{
			name: namePointMultG2,
			bs:   testing.Benchmark(func(b *testing.B) { benchScalarMult(b, curve.TypeG2) }),
		},
	)
	results = append(results,
		ExecResult{
			name: namePointMultGT,
			bs:   testing.Benchmark(func(b *testing.B) { benchScalarMult(b, curve.TypeGT) }),
		},
	)
	// point scalar div
	results = append(results,
		ExecResult{
			name: namePointDivG1,
			bs:   testing.Benchmark(func(b *testing.B) { benchScalarDiv(b, curve.TypeG1) }),
		},
	)
	results = append(results,
		ExecResult{
			name: namePointDivG2,
			bs:   testing.Benchmark(func(b *testing.B) { benchScalarDiv(b, curve.TypeG2) }),
		},
	)
	results = append(results,
		ExecResult{
			name: namePointDivGT,
			bs:   testing.Benchmark(func(b *testing.B) { benchScalarDiv(b, curve.TypeGT) }),
		},
	)
	// point pair
	results = append(results,
		ExecResult{
			name: namePointPair,
			bs:   testing.Benchmark(func(b *testing.B) { benchPair(b) }),
		},
	)
	// generate key
	results = append(results,
		ExecResult{
			name: nameGenerateKey,
			bs:   testing.Benchmark(func(b *testing.B) { benchGenerateKey(b) }),
		},
	)
	// generate re_key
	results = append(results,
		ExecResult{
			name: nameGenerateReKey,
			bs:   testing.Benchmark(func(b *testing.B) { benchGenerateReKey(b) }),
		},
	)
	// re_encrypt
	results = append(results,
		ExecResult{
			name: nameReEncrypt,
			bs:   testing.Benchmark(func(b *testing.B) { benchReEncrypt(b) }),
		},
	)
	// encrypt
	results = append(results,
		ExecResult{
			name:  nameEncrypt,
			bs:    testing.Benchmark(func(b *testing.B) { benchEncrypt(b, 1_000_000) }),
			extra: map[string]interface{}{"msg_size": 1_000_000},
		},
	)
	// decrypt
	results = append(results,
		ExecResult{
			name:  nameDecrypt,
			bs:    testing.Benchmark(func(b *testing.B) { benchDecrypt(b, 1_000_000) }),
			extra: map[string]interface{}{"msg_size": 1_000_000},
		},
	)
	// share message
	results = append(results,
		ExecResult{
			name: nameShareMessage,
			bs: testing.Benchmark(func(b *testing.B) {
				benchShareMessage(b, 1_000_000, 10)
			}),
			extra: map[string]interface{}{"msg_size": 1_000_000, "msg_count": 10},
		},
	)

	return results
}

func ShareMessageResults(sizeMin, sizeMax, sizeRatio, countMin, countMax, countStep int64) (
	results []ExecResult, sizes []int64, counts []int64) {
	for size := sizeMin; size <= sizeMax; size *= sizeRatio {
		sizes = append(sizes, size)
	}
	for count := countMin; count <= countMax; count += countStep {
		counts = append(counts, count)
	}
	//msgPreAlloc, cipherPreAlloc := make([]byte, sizeMax), make([]byte, sizeMax+4096)
	for size := sizeMin; size <= sizeMax; size *= sizeRatio {
		for count := countMin; count <= countMax; count += countStep {
			results = append(results,
				ExecResult{
					name:  nameShareMessage,
					bs:    testing.Benchmark(func(b *testing.B) { benchShareMessage(b, size, count) }),
					extra: map[string]interface{}{"msg_size": size, "msg_count": count},
				},
			)
			fmt.Println(size, count, results[len(results)-1])
		}
	}
	return
}

func Comparison(sizeMin, sizeMax, sizeRatio, countMin, countMax, countStep int64) (
	ourWork, otPaper, yaoGang []float64, sizes []int64, counts []int64) {
	for size := sizeMin; size <= sizeMax; size *= sizeRatio {
		sizes = append(sizes, size)
	}
	for count := countMin; count <= countMax; count += countStep {
		counts = append(counts, count)
	}
	// generate key
	kg := ExecResult{
		name: nameGenerateKey,
		bs:   testing.Benchmark(func(b *testing.B) { benchGenerateKey(b) }),
	}
	kgT := kg.MsPerOp()
	fmt.Println(kg.String())
	// generate re_key
	rkg := ExecResult{
		name: nameGenerateReKey,
		bs:   testing.Benchmark(func(b *testing.B) { benchGenerateReKey(b) }),
	}
	rkgT := rkg.MsPerOp()
	fmt.Println(rkg.String())
	// re_encrypt
	re := ExecResult{
		name: nameReEncrypt,
		bs:   testing.Benchmark(func(b *testing.B) { benchReEncrypt(b) }),
	}
	reT := re.MsPerOp()
	fmt.Println(re.String())
	// encrypt 32 bytes
	enc32 := ExecResult{
		name:  nameEncrypt,
		bs:    testing.Benchmark(func(b *testing.B) { benchAESEncrypt(b, 32) }),
		extra: map[string]interface{}{"msg_size": 32},
	}
	enc32T := enc32.MsPerOp()
	fmt.Println(enc32.String())
	// encrypt size bytes
	encTs := make([]float64, 0, len(sizes))
	for i := range sizes {
		enc := ExecResult{
			name:  nameEncrypt,
			bs:    testing.Benchmark(func(b *testing.B) { benchAESEncrypt(b, sizes[i]) }),
			extra: map[string]interface{}{"msg_size": sizes[i]},
		}
		encT := enc.MsPerOp()
		encTs = append(encTs, encT)
		fmt.Println(enc.String())
	}
	for i, size := range sizes {
		for count := countMin; count <= countMax; count += countStep {
			ourWorkT := (kgT + rkgT + reT) * float64(count)
			otPaperT := (kgT + encTs[i]) * float64(count)
			yaoGangT := (kgT*2 + enc32T + encTs[i]) * float64(count)
			ourWork = append(ourWork, ourWorkT)
			otPaper = append(otPaper, otPaperT)
			yaoGang = append(yaoGang, yaoGangT)
			fmt.Printf("size: %d, count: %d, out_work: %.3f, ot: %.3f, yao_gang: %.3f\n",
				size, count, ourWorkT, otPaperT, yaoGangT)
		}
	}
	return
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
