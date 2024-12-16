package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

const Th = 6

func SingleHash(in, out chan interface{}) {
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

		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func(out chan interface{}, leftCh chan string, rightCh chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			out <- <-leftCh + "~" + <-rightCh
		}(out, leftValueCh, rightValueCh, wg)
		defer wg.Wait()
	}
}

func MultiHash(in, out chan interface{}) {
	for ch := range in {
		counter := make(chan string, Th)
		ind := make(chan int, Th)

		wg := &sync.WaitGroup{}

		//Записывам в канал counter значения crc32(th+data),
		//а в канал ind значения th
		for th := 0; th < Th; th++ {
			wg.Add(1)
			go func(counter chan string, wg *sync.WaitGroup) {
				defer wg.Done()
				counter <- DataSignerCrc32(fmt.Sprintf("%v", th) + fmt.Sprintf("%v", ch))
				ind <- th
			}(counter, wg)
		}

		go func() {
			wg.Wait()
			close(counter)
			close(ind)
		}()

		wg1 := &sync.WaitGroup{}

		//Сортируем значения crc32(th+data) по th и пишем результат канал out
		wg1.Add(1)
		go func(wg1 *sync.WaitGroup) {
			defer wg1.Done()

			counters := make(map[int]string)
			for ch := range counter {
				counters[<-ind] = ch
			}

			keys := make([]int, 0, Th)
			for k := range counters {
				keys = append(keys, k)
			}
			sort.Ints(keys)

			var result string
			for _, val := range keys {
				result = result + counters[val]
			}

			out <- result
		}(wg1)
		defer wg1.Wait()
	}
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

// func MultiHash(in, out chan interface{}) {
// 	th0Ch := make(chan string, 1)
// 	th1Ch := make(chan string, 1)
// 	th2Ch := make(chan string, 1)
// 	th3Ch := make(chan string, 1)
// 	th4Ch := make(chan string, 1)
// 	th5Ch := make(chan string, 1)

// 	for ch := range in {
// 		go func(th0Ch chan string) { th0Ch <- DataSignerCrc32(fmt.Sprintf("%v", 0) + fmt.Sprintf("%v", ch)) }(th0Ch)
// 		go func(th1Ch chan string) { th1Ch <- DataSignerCrc32(fmt.Sprintf("%v", 1) + fmt.Sprintf("%v", ch)) }(th1Ch)
// 		go func(th2Ch chan string) { th2Ch <- DataSignerCrc32(fmt.Sprintf("%v", 2) + fmt.Sprintf("%v", ch)) }(th2Ch)
// 		go func(th3Ch chan string) { th3Ch <- DataSignerCrc32(fmt.Sprintf("%v", 3) + fmt.Sprintf("%v", ch)) }(th3Ch)
// 		go func(th4Ch chan string) { th4Ch <- DataSignerCrc32(fmt.Sprintf("%v", 4) + fmt.Sprintf("%v", ch)) }(th4Ch)
// 		go func(th5Ch chan string) { th5Ch <- DataSignerCrc32(fmt.Sprintf("%v", 5) + fmt.Sprintf("%v", ch)) }(th5Ch)

// 		wg := &sync.WaitGroup{}
// 		wg.Add(1)
// 		go func(out chan interface{}, wg *sync.WaitGroup) {
// 			defer wg.Done()
// 			out <- <-th0Ch + <-th1Ch + <-th2Ch + <-th3Ch + <-th4Ch + <-th5Ch
// 		}(out, wg)
// 		defer wg.Wait()
// 	}
// }
