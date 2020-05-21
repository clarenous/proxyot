package bench_test

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/clarenous/proxyot/bench"
)

func TestExecutionResults(t *testing.T) {
	results := bench.ExecutionResults()
	for _, res := range results {
		fmt.Println(res)
	}
}

type ShareMessageResults struct {
	SizeMin     int64     `json:"size_min"`
	SizeMax     int64     `json:"size_max"`
	SizeRatio   int64     `json:"size_ratio"`
	CountMin    int64     `json:"count_min"`
	CountMax    int64     `json:"count_max"`
	CountStep   int64     `json:"count_step"`
	Sizes       []int64   `json:"sizes"`
	Counts      []int64   `json:"counts"`
	ExecMsTimes []float64 `json:"exec_ms_times"`
}

func TestShareMessageResults(t *testing.T) {
	var sizeMin, sizeMax, sizeRatio int64 = 1_000, 1_000_000, 10
	var countMin, countMax, countStep int64 = 10, 110, 10

	results, sizes, counts := bench.ShareMessageResults(sizeMin, sizeMax, sizeRatio, countMin, countMax, countStep)
	execMsTimes := make([]float64, len(results))
	for i := range execMsTimes {
		execMsTimes[i] = results[i].MsPerOp()
	}

	// generate csv output
	csvFile, err := os.Create("share_message_results.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer csvFile.Close()
	csvWriter := csv.NewWriter(csvFile)
	var csvLine = make([]string, len(counts)+1)
	csvLine[0] = ""
	for i := range counts {
		csvLine[i+1] = fmt.Sprintf("%d", counts[i])
	}
	if err := csvWriter.Write(csvLine); err != nil {
		t.Fatal(err)
	}
	for i, size := range sizes {
		csvLine[0] = fmt.Sprintf("%d", size)
		for j := range counts {
			csvLine[j+1] = fmt.Sprintf("%.3f", execMsTimes[i*len(counts)+j])
		}
		if err := csvWriter.Write(csvLine); err != nil {
			t.Fatal(err)
		}
	}
	csvWriter.Flush()

	// generate json output
	shareMsgResults := &ShareMessageResults{
		SizeMin:     sizeMin,
		SizeMax:     sizeMax,
		SizeRatio:   sizeRatio,
		CountMin:    countMin,
		CountMax:    countMax,
		CountStep:   countStep,
		Sizes:       sizes,
		Counts:      counts,
		ExecMsTimes: execMsTimes,
	}
	jsonFile, err := os.Create("share_message_results.json")
	if err != nil {
		t.Fatal(err)
	}
	defer jsonFile.Close()
	jsonEncoder := json.NewEncoder(jsonFile)
	jsonEncoder.SetIndent("", "  ")
	if err = jsonEncoder.Encode(shareMsgResults); err != nil {
		t.Fatal(err)
	}
}

type AliceCommunicationCostResults struct {
	SizeMin            int64   `json:"size_min"`
	SizeMax            int64   `json:"size_max"`
	SizeRatio          int64   `json:"size_ratio"`
	CountMin           int64   `json:"count_min"`
	CountMax           int64   `json:"count_max"`
	CountStep          int64   `json:"count_step"`
	Sizes              []int64 `json:"sizes"`
	Counts             []int64 `json:"counts"`
	CommunicationCosts []int64 `json:"communication_costs"`
}

func TestAliceCommunicationCostResults(t *testing.T) {
	var sizeMin, sizeMax, sizeRatio int64 = 1_000, 1_000_000_000, 10
	var countMin, countMax, countStep int64 = 0, 110, 10

	costs, sizes, counts := bench.AliceCommunicationCostResults(sizeMin, sizeMax, sizeRatio, countMin, countMax, countStep)
	// generate json output
	costResults := &AliceCommunicationCostResults{
		SizeMin:            sizeMin,
		SizeMax:            sizeMax,
		SizeRatio:          sizeRatio,
		CountMin:           countMin,
		CountMax:           countMax,
		CountStep:          countStep,
		Sizes:              sizes,
		Counts:             counts,
		CommunicationCosts: costs,
	}
	jsonFile, err := os.Create("alice_transmission_costs.json")
	if err != nil {
		t.Fatal(err)
	}
	defer jsonFile.Close()
	jsonEncoder := json.NewEncoder(jsonFile)
	jsonEncoder.SetIndent("", "  ")
	if err = jsonEncoder.Encode(costResults); err != nil {
		t.Fatal(err)
	}
}

type ComparisonResults struct {
	SizeMin   int64     `json:"size_min"`
	SizeMax   int64     `json:"size_max"`
	SizeRatio int64     `json:"size_ratio"`
	CountMin  int64     `json:"count_min"`
	CountMax  int64     `json:"count_max"`
	CountStep int64     `json:"count_step"`
	Sizes     []int64   `json:"sizes"`
	Counts    []int64   `json:"counts"`
	OurWork   []float64 `json:"our_work"`
	Ot        []float64 `json:"ot"`
	YaoGang   []float64 `json:"yao_gang"`
}

func TestComparisonResults(t *testing.T) {
	var sizeMin, sizeMax, sizeRatio int64 = 1_000, 1_000_000_000, 10
	var countMin, countMax, countStep int64 = 10, 110, 10

	ourWork, otPaper, yaoGang, sizes, counts := bench.Comparison(sizeMin, sizeMax, sizeRatio, countMin, countMax, countStep)
	// generate json output
	costResults := &ComparisonResults{
		SizeMin:   sizeMin,
		SizeMax:   sizeMax,
		SizeRatio: sizeRatio,
		CountMin:  countMin,
		CountMax:  countMax,
		CountStep: countStep,
		Sizes:     sizes,
		Counts:    counts,
		OurWork:   ourWork,
		Ot:        otPaper,
		YaoGang:   yaoGang,
	}
	jsonFile, err := os.Create("comparison_results.json")
	if err != nil {
		t.Fatal(err)
	}
	defer jsonFile.Close()
	jsonEncoder := json.NewEncoder(jsonFile)
	jsonEncoder.SetIndent("", "  ")
	if err = jsonEncoder.Encode(costResults); err != nil {
		t.Fatal(err)
	}
}
