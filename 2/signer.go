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

	chanLeft := make(chan string)
	chanRight := make(chan string)

	for ch := range in {
		// DataSignerMd5 Считаем вне горутин, чтобы не перегреваться
		// Параллелим левую и правую часть, так как DataSignerCrc32, считается 1 сек
		md5 := DataSignerMd5(fmt.Sprintf("%v", ch))

		go func(data string, chanLeft chan string) {
			chanLeft <- DataSignerCrc32(data)
		}(fmt.Sprintf("%v", ch), chanLeft)

		go func(data string, chanRight chan string) {
			chanRight <- DataSignerCrc32(data)
		}(md5, chanRight)

		wg.Add(1)
		go func(out chan interface{}, chanLeft chan string, chanRight chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			out <- <-chanLeft + "~" + <-chanRight
		}(out, chanLeft, chanRight, wg)
	}

	defer wg.Wait()
}

func MultiHash(in, out chan interface{}) {

	wg := &sync.WaitGroup{}

	channels := make([]chan string, Th) // Создаем слайс каналов размером Th
	for i := range channels {
		channels[i] = make(chan string)
	}

	for ch := range in {
		for th := 0; th < Th; th++ {
			go func(th string, data string, chanTh chan string) {
				chanTh <- DataSignerCrc32(th + data)
			}(fmt.Sprintf("%v", th), fmt.Sprintf("%v", ch), channels[th])
		}

		res := make([]string, Th)

		wg.Add(1)
		go func(out chan interface{}, wg *sync.WaitGroup, res []string) {
			defer wg.Done()
			for i := range channels {
				res[i] = <-channels[i]
			}
			out <- strings.Join(res, "")
		}(out, wg, res)
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
