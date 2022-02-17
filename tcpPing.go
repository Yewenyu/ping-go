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
	var wg sync.WaitGroup
	count := len(addrs)
	wg.Add(count)

	var lock sync.Mutex
	var resultm = make([]PingResults, count)
	var gc = count / groupCount
	var currentPingCount = 0
	for i, addr := range addrs {
		arr := strings.Split(addr, ":")
		host := arr[0]
		port, err := strconv.Atoi(arr[1])
		if err != nil {
			continue
		}
		lock.Lock()
		currentPingCount += 1
		lock.Unlock()
		go func(i int) {
			r := StartTcpPing(host, port, pingCount, timeout, interval, checkAlive)
			
			lock.Lock()
			resultm[i] = r
			currentPingCount -= 1
			lock.Unlock()
			wg.Done()
			debug.FreeOSMemory()
		}(i)
		for{
			lock.Lock()
			if currentPingCount < gc{
				lock.Unlock()
				break
			}
			lock.Unlock()
		}
		
		
	}

	
	wg.Wait()
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
