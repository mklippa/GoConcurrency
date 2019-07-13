package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// сюда писать код
func main() {
	fmt.Println(CombineResults(MultiHash(SingleHash("0")), MultiHash(SingleHash("1"))))
}

func ExecutePipeline(jobs ...job) {

}

func CombineResults(results ...string) string {
	sort.Strings(results)
	return strings.Join(results, "_")
}

func SingleHash(data string) string {
	return DataSignerCrc32(data) + "~" + DataSignerCrc32(DataSignerMd5(data))
}

func MultiHash(data string) string {
	var result string
	for i := 0; i <= 5; i++ {
		result += DataSignerCrc32(strconv.Itoa(i) + data)
	}
	return result
}
