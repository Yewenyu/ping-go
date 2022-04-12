package goPing

import (
	"context"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	ping "github.com/Yewenyu/ping-go/ping"
	tcptest "github.com/Yewenyu/ping-go/tcptest"
)

type TcpTestResult struct{
	Alive []tcptest.Result
	Dead []tcptest.Result
	Timeout []tcptest.Result
}

var ctxCancels map[int]context.CancelFunc = make(map[int]context.CancelFunc)
func CheckTCPAliveCancel(tag int){
	cancel := ctxCancels[tag]
	if cancel != nil{
		cancel()
		
	}
}
func CheckTCPAlive(addrs []string,timeout int,maxCount int,tag int,send64Bytes bool,handleResult func(result PingResults)) () {
	ctx := context.Background()
	currentCtx,cancel := context.WithCancel(ctx)
	ctxCancels[tag] = cancel
	count := len(addrs)
	tcptest.Send64Bytes = send64Bytes
	
	var maxC = maxCount
	var groupChan = make(chan PingResults)
	var endChan = make(chan int,maxC)
	var wg sync.WaitGroup
	wg.Add(1)
	go func(){
		currentResultCount := 0
		for{
			result := <- groupChan
			handleResult(result)
			currentResultCount += 1
			if currentResultCount == count{
				wg.Done()
				break
			}
		}
	}()
	var currentPingCount = 0
	for _, addr := range addrs {
		currentPingCount += 1
		nCtx,_ := context.WithTimeout(currentCtx,time.Duration(timeout) * time.Millisecond)
		go func(a string) {
			s,d := tcptest.CheckTCP64(nCtx,a)
			arr := strings.Split(a, ":")
			host := arr[0]
			port, _ := strconv.Atoi(arr[1])
			errorType := 0
			t := int(d) / 1000 / 1000
			select {
			case <- currentCtx.Done():
				errorType = 3
			default:
				if !s{
					if t >= timeout{
						errorType = 2
					}else{
						errorType = 1
					}
					
				}
			}
			endChan <- 1
			groupChan <- PingResults{
				Host: host,
				Port: port,
				Time: t,
				ErrType: errorType,
			}
		}(addr)
		if currentPingCount > maxCount{
			currentPingCount -= <-endChan
		}
	}
	wg.Wait()
	close(endChan)
	close(groupChan)
	delete(ctxCancels,tag)
	debug.FreeOSMemory()
	runtime.GC()
	return
}

func StartTcpTest(addrs []string,timeout int,maxCount int) TcpTestResult{
	ctx := context.Background()
	ctx1,_ := context.WithTimeout(ctx,time.Duration(timeout)*time.Millisecond)
	
	as,ds,ts := tcptest.CheckTCPAlive(ctx1,addrs)
	

	return TcpTestResult{
		Alive: as,
		Dead: ds,
		Timeout: ts,
	}
}

func StartTcpPings(addrs []string, pingCount int, timeout int, interval int,groupCount int, checkAlive bool) []PingResults {
	
	count := len(addrs)
	var resultm = make([]PingResults,0)
	var sC = groupCount
	if sC == 0{
		sC = 1
	}else if sC > count{
		sC = 1
	}
	var gc = count / sC
	var groupChan = make(chan PingResults)
	var endChan = make(chan int,gc)
	var wg sync.WaitGroup
	wg.Add(1)
	go func(){
		for{
			result := <- groupChan
			resultm = append(resultm, result)
			if len(resultm) == count{
				wg.Done()
				break
			}
		}
	}()
	var currentPingCount = 0
	for i, addr := range addrs {
		arr := strings.Split(addr, ":")
		host := arr[0]
		port, err := strconv.Atoi(arr[1])
		if err != nil {
			continue
		}
		currentPingCount += 1
		go func(i int) {
			r := StartTcpPing(host, port, pingCount, timeout, interval, checkAlive)
			endChan <- 1
			groupChan <- r
			debug.FreeOSMemory()
		}(i)
		if currentPingCount > gc{
			currentPingCount -= <-endChan
		}
	}
	wg.Wait()
	close(endChan)
	close(groupChan)
	debug.FreeOSMemory()
	runtime.GC()
	return resultm
}
func StartTcpPing(host string, port int, pingCount int, timeout int, interval int, checkAlive bool) PingResults {

	target := ping.Target{
		Timeout:  time.Duration(timeout) * time.Second,
		Interval: time.Duration(interval) * time.Millisecond,
		Host:     host,
		Port:     port,
		Counter:  pingCount,
		Protocol: ping.TCP,
	}
	pinger := ping.NewTCPing()
	pinger.CheckAlive = checkAlive
	pinger.SetTarget(&target)
	pingerDone := pinger.Start()
	select {
	case <-pingerDone:
		break
	}
	result := pinger.Result()

	r := PingResults{
		Host: host,
		Port: port,
		Time: int(result.Avg().Milliseconds()),
		Loss: (1 - float64(result.SuccessCounter)/float64(result.Counter)) * 100,
		ErrType: pinger.ErrType,
	}
	return r
}
