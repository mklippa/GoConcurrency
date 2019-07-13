package main

import (
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код
func main() {
	runtime.GOMAXPROCS(0)
	// fmt.Println(CombineResults(MultiHash(SingleHash("0")), MultiHash(SingleHash("1"))))
}

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{}, 0)
	out := make(chan interface{}, 0)
	for _, j := range jobs {
		wg.Add(1)
		go func(worker job, inCh, outCh chan interface{}) {
			worker(inCh, outCh)
			runtime.Gosched()
			wg.Done()
		}(j, in, out)
	}
	wg.Done()
}

var CombineResults job = func(in, out chan interface{}) {
	result := make([]string, 0)
	for {
		select {
		case d := <-in:
			data, _ := d.(string)
			result = append(result, data)
		}
	}
	sort.Strings(result)
	out <- strings.Join(result, "_")
}

var SingleHash job = func(in, out chan interface{}) {
	for d := range in {
		data, _ := d.(string)
		go func(data string) {
			out <- DataSignerCrc32(data) + "~" + DataSignerCrc32(DataSignerMd5(data))
		}(data)
	}
}

var MultiHash job = func(in, out chan interface{}) {
	for d := range in {
		data, _ := d.(string)
		var result string
		for i := 0; i <= 5; i++ {
			result += DataSignerCrc32(strconv.Itoa(i) + data)
		}
		out <- result
	}
}
