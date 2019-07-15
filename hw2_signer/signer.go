package main

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// сюда писать код
func main() {
	runtime.GOMAXPROCS(0)
	// fmt.Println(CombineResults(MultiHash(SingleHash("0")), MultiHash(SingleHash("1"))))
}

func ExecutePipeline(jobs ...job) {
	gen := func() chan interface{} {
		out := make(chan interface{})
		go func() {
			jobs[0](nil, out)
			close(out)
		}()
		return out
	}

	in := gen()

	calc := func(i int, in chan interface{}) chan interface{} {
		out := make(chan interface{})
		go func() {
			jobs[i](in, out)
			close(out)
		}()
		return out
	}

	outs := make([]chan interface{}, 0)
	for i := 0; i < 100; i++ {
		out := in
		for j := 1; j < len(jobs)-1; j++ {
			out = calc(j, out)
		}
		outs = append(outs, out)
	}

	merge := func(cs ...chan interface{}) chan interface{} {
		var wg sync.WaitGroup
		out := make(chan interface{})

		// Start an output goroutine for each input channel in cs.  output
		// copies values from c to out until c is closed, then calls wg.Done.
		output := func(c <-chan interface{}) {
			for n := range c {
				out <- n
			}
			wg.Done()
		}
		wg.Add(len(cs))
		for _, c := range cs {
			go output(c)
		}

		// Start a goroutine to close out once all the output goroutines are
		// done.  This must start after the wg.Add call.
		go func() {
			wg.Wait()
			close(out)
		}()
		return out
	}

	res := merge(outs...)
	jobs[len(jobs)-1](res, nil)
}

var CombineResults job = func(in, out chan interface{}) {
	result := make([]string, 0)
	for val := range in {
		data, _ := val.(string)
		result = append(result, data)
	}
	start := time.Now()
	sort.Strings(result)
	end := time.Since(start)
	fmt.Println(end)
	out <- strings.Join(result, "_")
}

var mu = &sync.Mutex{}

var SingleHash job = func(in, out chan interface{}) {
	for val := range in {
		start := time.Now()
		data := strconv.Itoa(val.(int))

		crc32md5 := make(chan string, 1)
		go func(res chan<- string) {
			mu.Lock()
			md5 := DataSignerMd5(data)
			mu.Unlock()
			res <- DataSignerCrc32(md5)
		}(crc32md5)
		result := DataSignerCrc32(data) + "~" + (<-crc32md5)

		end := time.Since(start)
		fmt.Println(data, "SingleHash: ", end, result)
		out <- result
	}
}

var MultiHash job = func(in, out chan interface{}) {
	for val := range in {
		start := time.Now()

		data, _ := val.(string)

		res := [6]string{}
		wg := &sync.WaitGroup{}
		for i := 0; i < 6; i++ {
			wg.Add(1)
			go func(i int) {
				res[i] = DataSignerCrc32(strconv.Itoa(i) + data)
				wg.Done()
			}(i)
		}
		wg.Wait()

		result := strings.Join(res[:], "")

		end := time.Since(start)
		fmt.Println(data, "MultiHash: ", end)
		out <- result
	}
}
