package goPing

import (
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	ping "github.com/Yewenyu/ping-go/ping"
)

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
