package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
)

type DataCollector struct {
	best    float64
	current float64
	window  []float64
	winSize int
	samples int
}

func NewDC(wsz int) *DataCollector {
	return &DataCollector{winSize: wsz, window: make([]float64, wsz)}
}

func (dc *DataCollector) rescan(first bool) {
	min := dc.window[0]
	for _, v := range dc.window {
		if v < min {
			min = v
		}
	}
	dc.current = min
	if min > dc.best || first {
		dc.best = min
	}
}

func (dc *DataCollector) AddSample(s float64) {
	replaced := dc.window[dc.samples%dc.winSize]
	dc.window[dc.samples%dc.winSize] = s
	dc.samples++
	if dc.samples < dc.winSize {
		return
	}
	if dc.samples == dc.winSize || replaced == dc.current || s < dc.current {
		dc.rescan(dc.samples == dc.winSize)
	}
}

func (dc DataCollector) Result() float64 {
	if dc.samples < dc.winSize {
		log.Fatal("too few samples")
	}
	return dc.best
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
		log.Fatal("ReadFile(dump.json): %v\n", err)
	}
	data := make([]map[string]interface{}, 0)
	err = json.Unmarshal(j, &data)
	if err != nil {
		log.Fatal("json.Unmarshal: %v\n", err)
	}
	dc := NewDC(nsamples)
	for _, v := range data {
		f, err := strconv.ParseFloat(v["value"].(string), 64)
		if err != nil {
			log.Fatal("parsing '%v' as float: %v", v["value"], err)
		}
		dc.AddSample(f)
	}

	fmt.Printf("%v\n", dc.Result())
}
