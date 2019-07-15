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
	in := make(chan interface{})
	out := make(chan interface{})

	wg := &sync.WaitGroup{}
	for i := 0; i < len(jobs); i++ {
		in, out = out, make(chan interface{})
		wg.Add(1)
		go func(i int, in, out chan interface{}) {
			jobs[i](in, out)
			close(out)
			wg.Done()
		}(i, in, out)
	}

	wg.Wait()

	time.Sleep(time.Second)
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
		data := strconv.Itoa(val.(int))
		go func() {
			start := time.Now()

			crc32md5 := make(chan string, 1)
			go func(res chan<- string) {
				mu.Lock()
				res <- DataSignerCrc32(DataSignerMd5(data))
				mu.Unlock()
			}(crc32md5)
			result := DataSignerCrc32(data) + "~" + (<-crc32md5)

			end := time.Since(start)
			fmt.Println(data, "SingleHash: ", end)
			out <- result
		}()
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
