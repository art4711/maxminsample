package main

import (
	"container/heap"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
)

type DataCollector struct {
	best    float64
	window  []float64
	heap    []int // pretend heap of the values in the window
	revheap []int // we need to know which window position corresponds to which heap position for heap.Fix
	samples int
}

func NewDC(wsz int) *DataCollector {
	dc := DataCollector{window: make([]float64, wsz), heap: make([]int, wsz), revheap: make([]int, wsz)}
	for k, _ := range dc.heap {
		dc.heap[k], dc.revheap[k] = k, k
	}
	return &dc
}

func (dc *DataCollector) AddSample(s float64) {
	pos := dc.samples % len(dc.window)
	dc.window[pos] = s
	dc.samples++
	if dc.samples < len(dc.window) {
		return
	}
	if dc.samples == len(dc.window) {
		heap.Init(dc)
		dc.best = dc.window[dc.heap[0]]
		return
	}

	heap.Fix(dc, dc.revheap[pos])
	if dc.window[dc.heap[0]] > dc.best {
		dc.best = dc.window[dc.heap[0]]
	}
}

func (dc DataCollector) Result() float64 {
	if dc.samples < len(dc.window) {
		log.Fatal("too few samples")
	}
	return dc.best
}

// Implements heap.Interface
func (dc DataCollector) Len() int {
	return len(dc.heap)
}

// Implements heap.Interface
func (dc DataCollector) Less(i, j int) bool {
	return dc.window[dc.heap[i]] < dc.window[dc.heap[j]]
}

// Implements heap.Interface
func (dc DataCollector) Swap(i, j int) {
	dc.heap[i], dc.heap[j] = dc.heap[j], dc.heap[i]
	dc.revheap[dc.heap[i]], dc.revheap[dc.heap[j]] = i, j
}

// Implements heap.Interface
func (dc DataCollector) Push(x interface{}) {
	panic("we never push to the heap")
}

// Implements heap.Interface
func (dc DataCollector) Pop() interface{} {
	panic("we never pop from the heap")
	return nil
}

func main() {
	flag.Parse()
	nsamples := 10
	if flag.NArg() > 0 {
		n, err := strconv.Atoi(flag.Arg(0))
		if err != nil {
			log.Fatal("usage: foo [size of sample window]")
		}
		nsamples = n
	}
	j, err := ioutil.ReadFile("dump.json")
	if err != nil {
		log.Fatalf("ReadFile(dump.json): %v\n", err)
	}
	data := make([]map[string]interface{}, 0)
	err = json.Unmarshal(j, &data)
	if err != nil {
		log.Fatalf("json.Unmarshal: %v\n", err)
	}
	dc := NewDC(nsamples)
	for _, v := range data {
		f, err := strconv.ParseFloat(v["value"].(string), 64)
		if err != nil {
			log.Fatalf("parsing '%v' as float: %v", v["value"], err)
		}
		dc.AddSample(f)
	}

	fmt.Printf("%v\n", dc.Result())
}
