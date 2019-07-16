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

var SingleHash job = func(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	internalOut := make(chan interface{}, 100)
	start := time.Now()
	for val := range in {
		data := strconv.Itoa(val.(int))
		wg.Add(1)
		go func() {
			crc32md5 := make(chan string, 1)
			go func(res chan<- string) {
				mu.Lock()
				res <- DataSignerCrc32(DataSignerMd5(data))
				mu.Unlock()
			}(crc32md5)

			internalOut <- DataSignerCrc32(data) + "~" + (<-crc32md5)
			wg.Done()
		}()
	}
	wg.Wait()
	close(internalOut)
	end := time.Since(start)
	fmt.Println(end)
	for r := range internalOut {
		out <- r
	}
}

var MultiHash job = func(in, out chan interface{}) {
	wg1 := &sync.WaitGroup{}
	start := time.Now()
	internalOut := make(chan interface{}, 100)
	for val := range in {
		data := val.(string)
		wg1.Add(1)
		go func() {
			res := [6]string{}
			wg2 := &sync.WaitGroup{}
			for i := 0; i < 6; i++ {
				wg2.Add(1)
				go func(i int) {
					res[i] = DataSignerCrc32(strconv.Itoa(i) + data)
					wg2.Done()
				}(i)
			}
			wg2.Wait()
			internalOut <- strings.Join(res[:], "")
			wg1.Done()
		}()
	}
	wg1.Wait()
	close(internalOut)
	end := time.Since(start)
	fmt.Println(end)
	for r := range internalOut {
		out <- r
	}
}
