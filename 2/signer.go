package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func SingleHash(data int) string {
	s := DataSignerCrc32(fmt.Sprintf("%s", data)) + "~" + DataSignerCrc32(DataSignerMd5(fmt.Sprintf("%s", data)))
	return s
}

func MultiHash(data string) string {
	var result string
	for th := 0; th <= 5; th++ {
		result += fmt.Sprintf("%s", DataSignerCrc32(fmt.Sprintf("%s", th)+data))
	}

	return result
}

func CombineResults(data []string) string {

	sort.Strings(data)
	return strings.Join(data, "_")
}

func ExecutePipeline(freeFlowJobs []job) {

}

func main() {

	inputData := []int{0, 1, 1, 2, 3, 5, 8}
	// inputData := []int{0, 1}
	combine := make([]string, 0, len(inputData))

	start := time.Now()

	for _, v := range inputData {
		sh := SingleHash(v)
		mh := MultiHash(sh)
		combine = append(combine, mh)
	}
	result := CombineResults(combine)

	end := time.Since(start)
	fmt.Println(end, result)

}
