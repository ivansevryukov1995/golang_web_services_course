package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func SingleHash(in, out chan interface{}) {
	leftValueCh := make(chan string, 1)
	rightValueCh := make(chan string, 1)

	for ch := range in {
		md5 := DataSignerMd5(fmt.Sprintf("%v", ch))

		go func(data string, leftCh chan string) { // Параллелим левую и правую часть
			leftCh <- DataSignerCrc32(data)
		}(fmt.Sprintf("%v", ch), leftValueCh)

		go func(data string, rightCh chan string) {
			rightCh <- DataSignerCrc32(md5)
		}(md5, rightValueCh)

		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func(out chan interface{}, leftCh chan string, rightCh chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			leftValue := <-leftCh
			rightValue := <-rightCh
			out <- leftValue + "~" + rightValue
		}(out, leftValueCh, rightValueCh, wg)
		defer wg.Wait()
	}
}

func MultiHash(in, out chan interface{}) {
	for ch := range in {
		wg := &sync.WaitGroup{}
		counter := make(chan string)
		var counters = map[int]string{}

		for th := 0; th < 6; th++ {
			wg.Add(1)
			go func(counter chan string, wg *sync.WaitGroup) {
				defer wg.Done()
				counter <- fmt.Sprintf("%d_", th) + DataSignerCrc32(fmt.Sprintf("%v", th)+fmt.Sprintf("%v", ch))
			}(counter, wg)
		}

		go func() {
			wg.Wait()
			close(counter)
		}()

		wg1 := &sync.WaitGroup{}
		wg1.Add(1)
		go func(out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			for count := range counter {
				i, err := strconv.Atoi(strings.Split(count, "_")[0])
				if err != nil {
					log.Fatal(err)
				}
				counters[i] = strings.Split(count, "_")[1]
			}

			keys := make([]int, 0, len(counters))

			for k := range counters {
				keys = append(keys, k)
			}
			sort.Ints(keys)

			var result string
			for _, val := range keys {
				result = result + counters[val]
			}

			out <- result

		}(out, wg1)
		defer wg1.Wait()
	}
}

func CombineResults(in, out chan interface{}) {
	var hashResults []string

	for hashResult := range in {
		hashResults = append(hashResults, (hashResult).(string))
	}

	sort.Strings(hashResults)

	out <- strings.Join(hashResults, "_")
}

func ExecutePipeline(hashSignJobs ...job) {
	in := make(chan interface{})
	wg := &sync.WaitGroup{}

	for _, jobItem := range hashSignJobs {
		wg.Add(1)
		out := make(chan interface{})
		go func(jobFunc job, in chan interface{}, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			defer close(out)
			jobFunc(in, out)
		}(jobItem, in, out, wg)

		in = out

		defer wg.Wait()
	}
}

// Код пока не самый лучший, потом может исправлю :)
// === RUN   TestByIlia
// collected 3
// collected 9
// collected 12
// --- PASS: TestByIlia (0.30s)
// === RUN   TestPipeline
// --- PASS: TestPipeline (0.01s)
// === RUN   TestSigner
// --- PASS: TestSigner (2.07s)
// PASS
// ok      2       2.565s
