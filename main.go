package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
	"flag"
	"sync"
)

var wg sync.WaitGroup
var (
	concurrency int = 1      // 并发数
	totalNumber int = 1      // 请求数(单个并发/协程)
	unixSocket  string = "/tmp/a.sock"
	url         string = "http://unix/rest/helloworld"
)
func init() { 
	flag.IntVar(&concurrency, "c", concurrency, "并发数")
	flag.IntVar(&totalNumber, "n", totalNumber, "请求数(单个并发/协程)")
    flag.StringVar(&unixSocket, "unix-socket", unixSocket, ".sock文件路径")
	flag.StringVar(&url, "url", url, "请求地址")

	// 解析参数
	flag.Parse()
}


type requestStatus struct {
	success bool 
	duration int64
 }

func clientRequest(chSuccess chan int) {
	// 一个客户端
	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", unixSocket)
			},
		},
	}
	
	// 该客户端的所有请求结果
	var allRequestStatus []requestStatus = make([]requestStatus, totalNumber)
	
	// 成功数
	var totalSuccess = 0

	for i := 0; i < totalNumber; i++ {
		pre := time.Now().UnixMicro()
		var response *http.Response
		var err error
		response, err = httpc.Get(url)

		// 判断请求是否成功
		if err != nil || response.StatusCode != 200 {
			allRequestStatus[i].success = false
		} else {
			allRequestStatus[i].success = true
			totalSuccess++
		}
		_ = response.Body

		allRequestStatus[i].duration = time.Now().UnixMicro() - pre
	}
	chSuccess <- totalSuccess
	for i := 0; i < totalNumber; i++ {
		fmt.Printf("status:%t, duration:%dus\n", allRequestStatus[i].success, allRequestStatus[i].duration)
	}
	wg.Done()
}

func doRequest() {
	chSuccess := make(chan int, concurrency)
	stratTime := time.Now().UnixMicro()
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go clientRequest(chSuccess)
	}
	wg.Wait()
	duration := time.Now().UnixMicro() - stratTime

	close(chSuccess)
	var totalSuccess = 0
	for i := range chSuccess {
		totalSuccess = totalSuccess + i
	}

	fmt.Printf("总用时：%.2fs\n", float64(duration) / 1000000)
	fmt.Printf("总请求数: %d\n", concurrency * totalNumber)
	fmt.Printf("成功请求数: %d\n", totalSuccess)
	fmt.Printf("QPS: %.2f\n", float64(concurrency * totalNumber) / (float64(duration) / 1000000))
}

func main() {
	fmt.Println("unix-socket: ", unixSocket)
	fmt.Println("url: ", url)

	doRequest()

	printResult()
}


func printResult() {
	fmt.Println("over")
}