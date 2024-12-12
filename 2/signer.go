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
	wg := &sync.WaitGroup{}

	for ch := range in {
		wg.Add(1)
		md5 := DataSignerMd5(fmt.Sprintf("%v", ch))

		go func(out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()

			rightValue := DataSignerCrc32(md5)

			leftValueCh := make(chan string, 1)
			go func(ch interface{}, leftValueCh chan string) {
				leftValueCh <- DataSignerCrc32(fmt.Sprintf("%v", ch))
			}(ch, leftValueCh)
			leftValue := <-leftValueCh

			out <- leftValue + "~" + rightValue
		}(out, wg)
		defer wg.Wait()
	}
}

func MultiHash(in, out chan interface{}) {

	for ch := range in {
		wg := &sync.WaitGroup{}
		counter := make(chan interface{})
		var counters = map[int]string{}

		wg.Add(1)
		go func(counter chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			for th := 0; th < 6; th++ {
				wg.Add(1)
				go func(counter chan interface{}, wg *sync.WaitGroup) {
					defer wg.Done()
					counter <- fmt.Sprintf("%d_", th) + DataSignerCrc32(fmt.Sprintf("%v", th)+fmt.Sprintf("%v", ch))
				}(counter, wg)
			}
		}(counter, wg)
		go func() {
			wg.Wait()
			close(counter)
		}()

		wg1 := &sync.WaitGroup{}
		wg1.Add(1)
		go func(out chan interface{}, wg1 *sync.WaitGroup) {
			defer wg1.Done()
			for count := range counter {
				i, err := strconv.Atoi(strings.Split(count.(string), "_")[0])
				if err != nil {
					log.Fatal(err)
				}
				// mu.Lock()
				counters[i] = strings.Split(count.(string), "_")[1]
				// mu.Unlock()
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
