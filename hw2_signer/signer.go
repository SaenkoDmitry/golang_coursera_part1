package main

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"
)

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	mainCh := make([]chan interface{}, len(jobs)+1)
	for i := range mainCh {
		mainCh[i] = make(chan interface{}, 100)
	}
	var in chan interface{}
	out := mainCh[0]
	ind := 1

	for i, j := range jobs {
		in = out
		out = mainCh[ind]
		ind++
		wg.Add(1)
		go func(i int, j job, in chan interface{}, out chan interface{}, wg *sync.WaitGroup) {
			j(in, out)
			if i == 0 {
				close(out)
			}
			wg.Done()
		}(i, j, in, out, wg)
		time.Sleep(time.Millisecond)
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for d := range in {
		wgIns := &sync.WaitGroup{}
		wg.Add(1)
		go func(d interface{}, wg *sync.WaitGroup) {
			data := fmt.Sprintf("%v", d)
			fmt.Println(d, "SingleHash data", data)

			mu.Lock()
			md5 := DataSignerMd5(data)
			mu.Unlock()
			fmt.Println(d, "SingleHash md5(data)", md5)

			var crc32Md5 string
			wgIns.Add(1)
			go func(md5 string) {
				crc32Md5 = DataSignerCrc32(md5)
				wgIns.Done()
			}(md5)

			var crc32 string
			wgIns.Add(1)
			go func(data string) {
				crc32 = DataSignerCrc32(data)
				wgIns.Done()
			}(data)
			wgIns.Wait()

			fmt.Println(d, "SingleHash crc32(md5(data))", crc32Md5)
			fmt.Println(d, "SingleHash crc32(data)", crc32)

			result := crc32 + "~" + crc32Md5
			out <- result
			fmt.Println(d, "SingleHash result", result)
			wg.Done()
		}(d, wg)
		runtime.Gosched()
	}
	wg.Wait()
	close(out)
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	for d := range in {
		wgIns := &sync.WaitGroup{}
		wg.Add(1)
		go func(d interface{}, wg *sync.WaitGroup) {
			data := fmt.Sprintf("%v", d)
			result := ""
			for i := 0; i < 6; i++ {
				wgIns.Add(1)
				go func(j int, data string, result *string, wg *sync.WaitGroup) {
					th := strconv.Itoa(j)
					crc32 := DataSignerCrc32(th + data)
					mu.Lock()
					*result += crc32
					mu.Unlock()
					fmt.Println(d, "MultiHash: crc32(th+step1))", th, crc32)
					wgIns.Done()
				}(i, data, &result, wgIns)
				time.Sleep(time.Millisecond)
			}
			wgIns.Wait()
			out <- result
			fmt.Println(d, "MultiHash: result", result)
			fmt.Println()
			wg.Done()
		}(d, wg)
		runtime.Gosched()
	}
	wg.Wait()
	close(out)
}

type ByRes []string

func (a ByRes) Len() int           { return len(a) }
func (a ByRes) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRes) Less(i, j int) bool { return a[i] < a[j] }

func CombineResults(in, out chan interface{}) {
	var sorted []string
	for d := range in {
		data := fmt.Sprintf("%v", d)
		sorted = append(sorted, data)
	}
	sort.Sort(ByRes(sorted))
	result := ""
	var del string
	for _, x := range sorted {
		result = result + del + x
		del = "_"
	}
	fmt.Println(result)
	out <- result
	close(out)
}
