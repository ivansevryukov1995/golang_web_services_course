package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

const Th = 6

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	leftValueCh := make(chan string, 1)
	rightValueCh := make(chan string, 1)

	for ch := range in {
		// DataSignerMd5 Считаем вне горутин, чтобы не перегреваться
		// Параллелим левую и правую часть, так как DataSignerCrc32, считается 1 сек
		md5 := DataSignerMd5(fmt.Sprintf("%v", ch))

		go func(data string, leftCh chan string) {
			leftCh <- DataSignerCrc32(data)
		}(fmt.Sprintf("%v", ch), leftValueCh)

		go func(data string, rightCh chan string) {
			rightCh <- DataSignerCrc32(md5)
		}(md5, rightValueCh)

		wg.Add(1)
		go func(out chan interface{}, leftCh chan string, rightCh chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			out <- <-leftCh + "~" + <-rightCh
		}(out, leftValueCh, rightValueCh, wg)

	}
	defer wg.Wait()
}

type ThCh struct {
	ind  int
	data string
}

func MultiHash(in, out chan interface{}) {

	wg := &sync.WaitGroup{}

	thCh0 := make(chan string, 1)
	thCh1 := make(chan string, 1)
	thCh2 := make(chan string, 1)
	thCh3 := make(chan string, 1)
	thCh4 := make(chan string, 1)
	thCh5 := make(chan string, 1)

	chanel := map[string]chan string{
		"thCh0": thCh0,
		"thCh1": thCh1,
		"thCh2": thCh2,
		"thCh3": thCh3,
		"thCh4": thCh4,
		"thCh5": thCh5,
	}

	for ch := range in {
		for th := 0; th < Th; th++ {
			go func(th string, ch string, thCh chan string) {
				thCh <- DataSignerCrc32(th + ch)
			}(fmt.Sprintf("%v", th), fmt.Sprintf("%v", ch), chanel["thCh"+fmt.Sprintf("%v", th)])
		}

		wg.Add(1)
		go func(out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			out <- <-thCh0 + <-thCh1 + <-thCh2 + <-thCh3 + <-thCh4 + <-thCh5
		}(out, wg)

	}
	defer wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	hashResults := make([]string, 0, MaxInputDataLen) // Заранее выделяем в памяти слайс для 100 элементов

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
