package main

import (
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// сюда писать код
func main() {
	runtime.GOMAXPROCS(0)
	// fmt.Println(CombineResults(MultiHash(SingleHash("0")), MultiHash(SingleHash("1"))))
}

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{}, 100)
	out := make(chan interface{}, 100)

	for i := 0; i < len(jobs); i++ {
		go func(i int, in, out chan interface{}) {
			in, out = out, make(chan interface{}, 100)
			jobs[i](in, out)
			close(out)
		}(i, in, out)
	}
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
	data, _ := (<-in).(string)
	out <- DataSignerCrc32(data) + "~" + DataSignerCrc32(DataSignerMd5(data))
}

var MultiHash job = func(in, out chan interface{}) {
	data, _ := (<-in).(string)
	var result string
	for i := 0; i <= 5; i++ {
		result += DataSignerCrc32(strconv.Itoa(i) + data)
	}
	out <- result
}
